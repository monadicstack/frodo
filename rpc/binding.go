package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Binder performs the work of taking all meaningful values from an incoming request (body,
// path params, query string) and applying them to a Go struct (likely the "XxxRequest" for
// your service method).
type Binder interface {
	Bind(req *http.Request, out interface{}) error
}

// WithBinder allows you to override a Gateway's default binding behavior with the custom behavior
// of your choice.
func WithBinder(binder Binder) GatewayOption {
	return func(gw *Gateway) {
		gw.Binder = binder
	}
}

// jsonBinder is the default gateway binder that uses encoding/json to apply body/path/query data
// to service request models.
type jsonBinder struct{}

func (b jsonBinder) Bind(req *http.Request, out interface{}) error {
	if err := b.bindBody(req, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	if err := b.bindQueryString(req, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	if err := b.bindPathParams(req, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	return nil
}

func (b jsonBinder) bindBody(req *http.Request, out interface{}) error {
	if req.Body == nil {
		return nil
	}
	if req.Body == http.NoBody {
		return nil
	}
	return json.NewDecoder(req.Body).Decode(out)
}

func (b jsonBinder) bindQueryString(req *http.Request, out interface{}) error {
	//fmt.Printf("+++++++++++++ QUERY: %v\n", req.URL.Query())
	err := b.bindValues(req.URL.Query(), out)
	if err != nil {
		return fmt.Errorf("bind query string: %w", err)
	}
	return nil
}

func (b jsonBinder) bindPathParams(req *http.Request, out interface{}) error {
	params := httprouter.ParamsFromContext(req.Context())
	values := url.Values{}
	for _, param := range params {
		values[param.Key] = []string{param.Value}
	}

	//fmt.Printf(">>>>> Bind values: %v\n", values)
	err := b.bindValues(values, out)
	if err != nil {
		return fmt.Errorf("bind path params: %w", err)
	}
	return nil
}

type jsonType int

const (
	jsonTypeNil    = jsonType(0)
	jsonTypeString = jsonType(1)
	jsonTypeNumber = jsonType(2)
	jsonTypeBool   = jsonType(3)
	jsonTypeObject = jsonType(4)
	jsonTypeArray  = jsonType(5)
)

func (b jsonBinder) toClosestJSONType(outValue reflect.Value, key []string) jsonType {
	keyLength := len(key)
	if keyLength < 1 {
		return jsonTypeNil
	}
	if outValue.Kind() != reflect.Struct {
		return jsonTypeNil
	}

	currentType := outValue.Type()
	currentJSONType := jsonTypeObject
	for i := 0; i < keyLength; i++ {
		finalSegment := i >= keyLength-1
		field, ok := lookupField(currentType, key[i])

		if !ok {
			return jsonTypeNil
		}

		currentType = field.Type
		currentJSONType = fieldToJSONType(field)

		// There are more segments left in the key (e.g. we're only on "bar" of the key "foo.bar.baz.blah"),
		// so update our cursor type so that we can continue to iterate deeper into the structure. If, however,
		// the value is not a struct or map, we can't have attributes on it, so we've got a mismatch between
		// the incoming data that thinks the structure is more rich than it is. In the example maybe "foo.bar"
		// is actually a string when you look at the Go models, so we can't possibly resolve the
		// attribute "baz" on a string.
		if !finalSegment && currentJSONType != jsonTypeObject {
			return jsonTypeNil
		}
	}
	return currentJSONType
}

func fieldToJSONType(field reflect.StructField) jsonType {
	switch field.Type.Kind() {
	case reflect.String:
		return jsonTypeString
	case reflect.Bool:
		return jsonTypeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return jsonTypeNumber
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return jsonTypeNumber
	case reflect.Float32, reflect.Float64:
		return jsonTypeNumber
	case reflect.Array, reflect.Slice:
		return jsonTypeArray
	case reflect.Map, reflect.Struct:
		return jsonTypeObject
	default:
		return jsonTypeNil
	}
}

func (b jsonBinder) bindValues(requestValues url.Values, out interface{}) error {
	outValue := reflect.Indirect(reflect.ValueOf(out))

	// To keep the logic more simple (but fast enough for most use cases), we will generate a separate
	// JSON representation of each value and run it through the JSON decoder. To make things a bit more
	// efficient, we'll re-use the buffer/reader and the JSON decoder for each value we unmarshal
	// ont the output.
	buf := &bytes.Buffer{}
	decoder := json.NewDecoder(buf)

	for key, value := range requestValues {
		keySegments := strings.Split(key, ".")

		// Follow the segments of the key and determine the JSON type of the last segment. So if you
		// are binding the key "foo.bar.baz", we'll look at the Go data type of the "baz" field once
		// we've followed the path "out.foo.bar". This will spit back our enum for the JSON data type
		// that will most naturally unmarshal to the Go type. So if the Go data type for the "baz" field
		// is uint16 then we'd expect this to return 'jsonTypeNumber'. If "baz" were a string then
		// we'd expect this to return 'jsonTypeString', and so on.
		valueType := b.toClosestJSONType(outValue, keySegments)

		// We didn't find a field path with that name (e.g. the key was "name" but there was no field called "name")
		if valueType == jsonTypeNil {
			//log.Printf("skipping param without equivalent field: %s", key)
			continue
		}
		// Maybe you provided "foo.bar.baz=4" and there was is a field at "out.foo.bar.baz", but it's
		// a struct of some kind, so "4" is not enough to properly bind it. Arrays/slices we'll handle
		// in a future version... maybe.
		if valueType == jsonTypeObject || valueType == jsonTypeArray {
			continue
		}

		// Convert the parameter "foo.bar.baz=4" into {"foo":{"bar":{"baz":4}}} so that the standard
		// JSON decoder can work its magic to apply that to 'out' properly.
		buf.Reset()
		for i, keySegment := range keySegments {
			finalSegment := i == len(keySegments)-1

			buf.WriteString(`{"`)
			buf.WriteString(keySegment)
			buf.WriteString(`":`)
			if finalSegment {
				b.writeBindingValueJSON(buf, value[0], valueType)
			}
		}
		for i := 0; i < len(keySegments); i++ {
			buf.WriteString("}")
		}

		if err := decoder.Decode(out); err != nil {
			return fmt.Errorf("unable to bind '%s': %w", key, err)
		}
	}
	return nil
}

// writeBindingValueJSON outputs the right-hand-side of the JSON we're going to use to try and bind
// this value. For instance, when the binder is creating the JSON {"name":"bob"} for the
// parameter "name=bob", this function determines that "bob" is supposed to be written as a string
// and will write `"bob"` to the buffer.
func (b jsonBinder) writeBindingValueJSON(buf *bytes.Buffer, value string, valueType jsonType) {
	switch valueType {
	case jsonTypeString:
		buf.WriteString(`"`)
		buf.WriteString(value)
		buf.WriteString(`"`)
	case jsonTypeNumber, jsonTypeBool:
		buf.WriteString(value)
	default:
		// Whether its a nil (unknown) or object type, the binder doesn't support that type
		// of value, so just write null to avoid binding anything if we can help it.
		buf.WriteString("null")
	}
}

var noField = reflect.StructField{}

func lookupField(structType reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.EqualFold(name, fieldMappingName(field)) {
			return field, true
		}
	}
	return noField, false
}

func fieldMappingName(field reflect.StructField) string {
	jsonTag, ok := field.Tag.Lookup("json")
	if !ok {
		return field.Name
	}
	// You're actually omitting this field, so don't return something that can be matched
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}

	// Parse the `json` tag to determine how the user has re-mapped the field.
	switch comma := strings.IndexRune(jsonTag, ','); comma {
	case -1:
		// e.g. `json:"firstName"`
		return jsonTag
	case 0:
		// e.g. `json:",omitempty"` (not remapped so use fields actual name)
		return field.Name
	default:
		// e.g. `json:"firstName,omitempty" (just use the remapped name)
		return jsonTag[0:comma]
	}
}

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
	"github.com/robsignorelli/frodo/internal/reflection"
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
	if err := b.BindBody(req, out); err != nil {
		return fmt.Errorf("error binding body: %w", err)
	}
	if err := b.BindQueryString(req, out); err != nil {
		return fmt.Errorf("error binding query string: %w", err)
	}
	if err := b.BindPathParams(req, out); err != nil {
		return fmt.Errorf("error binding path params: %w", err)
	}
	return nil
}

func (b jsonBinder) BindBody(req *http.Request, out interface{}) error {
	if req.Body == nil {
		return nil
	}
	if req.Body == http.NoBody {
		return nil
	}
	return json.NewDecoder(req.Body).Decode(out)
}

func (b jsonBinder) BindQueryString(req *http.Request, out interface{}) error {
	return b.bindValues(req.URL.Query(), out)
}

func (b jsonBinder) BindPathParams(req *http.Request, out interface{}) error {
	params := httprouter.ParamsFromContext(req.Context())
	values := url.Values{}
	for _, param := range params {
		values[param.Key] = []string{param.Value}
	}
	return b.bindValues(values, out)
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
		valueType := b.keyToJSONType(outValue, keySegments)

		// We didn't find a field path with that name (e.g. the key was "name" but there was no field called "name")
		if valueType == jsonTypeNil {
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
		b.writeValueJSON(buf, keySegments, value[0], valueType)

		// Now that we have a close-enough JSON representation of your parameter, let the standard
		// JSON decoder do its magic.
		if err := decoder.Decode(out); err != nil {
			return fmt.Errorf("unable to bind value '%s'='%s': %w", key, value[0], err)
		}
	}
	return nil
}

// writeValueJSON accepts the decomposed parameter key (e.g. "foo.bar.baz") and the raw string value (e.g. "moo")
// and writes JSON to the buffer which can be used in standard JSON decoding/unmarshaling to apply the value
// to the out object (e.g. `{"foo":{"bar":{"baz":"moo"}}}`).
func (b jsonBinder) writeValueJSON(buf *bytes.Buffer, keySegments []string, value string, valueType jsonType) {
	for _, keySegment := range keySegments {
		buf.WriteString(`{"`)
		buf.WriteString(keySegment)
		buf.WriteString(`":`)
	}
	b.writeBindingValueJSON(buf, value, valueType)
	for i := 0; i < len(keySegments); i++ {
		buf.WriteString("}")
	}
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

// jsonType describes all of the possible JSON data types.
type jsonType int

const (
	jsonTypeNil    = jsonType(0)
	jsonTypeString = jsonType(1)
	jsonTypeNumber = jsonType(2)
	jsonTypeBool   = jsonType(3)
	jsonTypeObject = jsonType(4)
	jsonTypeArray  = jsonType(5)
)

func (b jsonBinder) keyToJSONType(outValue reflect.Value, key []string) jsonType {
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
		field, ok := reflection.FindField(currentType, key[i])
		if !ok {
			return jsonTypeNil
		}

		currentType = field.Type
		currentJSONType = b.fieldToJSONType(field)

		// There are more segments left in the key (e.g. we're only on "bar" of the key "foo.bar.baz.blah"),
		// so update our cursor type so that we can continue to iterate deeper into the structure. If, however,
		// the value is not a struct or map, we can't have attributes on it, so we've got a mismatch between
		// the incoming data that thinks the structure is more rich than it is. In the example maybe "foo.bar"
		// is actually a string when you look at the Go models, so we can't possibly resolve the
		// attribute "baz" on a string.
		finalSegment := i >= keyLength-1
		if !finalSegment && currentJSONType != jsonTypeObject {
			return jsonTypeNil
		}
	}
	return currentJSONType
}

// fieldToJSONType looks at the Go type of some field on a struct and returns the JSON data type
// that will most likely unmarshal to that field w/o an error.
func (b jsonBinder) fieldToJSONType(field reflect.StructField) jsonType {
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

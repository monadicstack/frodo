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
	"github.com/monadicstack/frodo/internal/reflection"
	"github.com/monadicstack/frodo/rpc/errors"
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
		if binder != nil {
			gw.Binder = binder
		}
	}
}

// jsonBinder is the default gateway binder that uses encoding/json to apply body/path/query data
// to service request models. This unifies the processing so that all 3 sources of values are unmarshaled
// using the same semantics. The goal is that whether a value comes in from the body or the path or the
// query string, if you supplied a custom UnmarshalJSON() function for a type, that will work.
//
// It assumes that the body is JSON, naturally. The path and query params, however, get massaged into
// JSON so that we can use the standard library's JSON package to unmarshal that data onto your out value.
// For example, let's assume that we have the following query string:
//
//     ?first=Bob&last=Smith&age=39&address.city=Seattle&enabled=true
//
// The jsonBinder will first create 5 separate JSON objects:
//
//     { "first": "Bob" }
//     { "last": "Smith" }
//     { "age": 39 }
//     { "address": { "city": "Seattle" } }     <-- notice how we handle nested values separated by "."
//     { "enabled": true }
//
// After generating each value, the jsonBinder will feed the massaged JSON to a 'json.Decoder' and standard
// JSON marshaling rules will overlay each one onto your 'out' value.
type jsonBinder struct{}

// jsonBindingContext carries our buffer/decoder context through all of the binding operations so
// that all values can share resources (e.g. binding the path params can piggy back off of the
// work of binding the query string).
type jsonBindingContext struct {
	buf     *bytes.Buffer
	decoder *json.Decoder
}

func (b jsonBinder) Bind(req *http.Request, out interface{}) error {
	if req == nil {
		return fmt.Errorf("unable to bind nil request")
	}
	// To keep the logic more simple (but fast enough for most use cases), we will generate a separate
	// JSON representation of each value and run it through the JSON decoder. To make things a bit more
	// efficient, we'll re-use the buffer/reader and the JSON decoder for each value we unmarshal
	// on the output. This way we only suffer one buffer allocation no matter how many values we handle.
	buf := &bytes.Buffer{}
	ctx := jsonBindingContext{
		buf:     buf,
		decoder: json.NewDecoder(buf),
	}

	if err := b.BindQueryString(ctx, req, out); err != nil {
		return fmt.Errorf("error binding query string: %w", err)
	}
	if err := b.BindBody(ctx, req, out); err != nil {
		return fmt.Errorf("error binding body: %w", err)
	}
	if err := b.BindPathParams(ctx, req, out); err != nil {
		return fmt.Errorf("error binding path params: %w", err)
	}
	return nil
}

// BindBody decodes the JSON body of the request onto the 'out' value.
func (b jsonBinder) BindBody(_ jsonBindingContext, req *http.Request, out interface{}) error {
	if req.Body == nil {
		return nil
	}
	if req.Body == http.NoBody {
		return nil
	}
	if req.Method != "POST" && req.Method != "PUT" && req.Method != "PATCH" {
		return nil // Only bind methods universally intended to have body data that affects the request.
	}
	return json.NewDecoder(req.Body).Decode(out)
}

// BindQueryString decodes all of the query string parameters onto the 'out' value. Each parameter will
// be converted to an equivalent JSON object and unmarshaled separately.
func (b jsonBinder) BindQueryString(ctx jsonBindingContext, req *http.Request, out interface{}) error {
	if req.URL == nil {
		return errors.BadRequest("request missing url")
	}
	return b.bindValues(ctx, req.URL.Query(), out)
}

// BindQueryString decodes all of the URL path parameters onto the 'out' value. Each parameter will
// be converted to an equivalent JSON object and unmarshaled separately.
func (b jsonBinder) BindPathParams(ctx jsonBindingContext, req *http.Request, out interface{}) error {
	params := httprouter.ParamsFromContext(req.Context())
	values := url.Values{}
	for _, param := range params {
		values[param.Key] = []string{param.Value}
	}
	return b.bindValues(ctx, values, out)
}

func (b jsonBinder) bindValues(ctx jsonBindingContext, requestValues url.Values, out interface{}) error {
	outValue := reflect.Indirect(reflect.ValueOf(out))

	for key, value := range requestValues {
		keySegments := strings.Split(key, ".")

		// Follow the segments of the key and determine the JSON type of the last segment. So if you
		// are binding the key "foo.bar.baz", we'll look at the Go data type of the "baz" field once
		// we've followed the path "out.foo.bar". This will spit back our enum for the JSON data type
		// that will most naturally unmarshal to the Go type. So if the Go data type for the "baz" field
		// is uint16 then we'd expect this to return 'jsonTypeNumber'. If "baz" were a string then
		// we'd expect this to return 'jsonTypeString', and so on.
		valueType := b.keyToJSONType(outValue, keySegments, value[0])

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
		ctx.buf.Reset()
		b.writeParamJSON(ctx.buf, keySegments, value[0], valueType)

		// Now that we have a close-enough JSON representation of your parameter, let the standard
		// JSON decoder do its magic.
		if err := ctx.decoder.Decode(out); err != nil {
			return fmt.Errorf("unable to bind value '%s'='%s': %w", key, value[0], err)
		}
	}
	return nil
}

// writeParamJSON accepts the decomposed parameter key (e.g. "foo.bar.baz") and the raw string value (e.g. "moo")
// and writes JSON to the buffer which can be used in standard JSON decoding/unmarshaling to apply the value
// to the out object (e.g. `{"foo":{"bar":{"baz":"moo"}}}`).
func (b jsonBinder) writeParamJSON(buf *bytes.Buffer, keySegments []string, value string, valueType jsonType) {
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

// keyToJSONType looks at your parameter key (e.g. "foo.bar.baz") and your value (e.g. "12345"), and
// indicates how we should format the value when creating binding JSON. It will use reflection to
// traverse the Go attributes foo, then bar, then baz, and return the most appropriate JSON type
// for the go type. For instance if "baz" is a uint16, the most appropriate jsonType is jsonTypeNumber.
func (b jsonBinder) keyToJSONType(outValue reflect.Value, key []string, value string) jsonType {
	keyLength := len(key)
	if keyLength < 1 {
		return jsonTypeNil
	}
	if outValue.Kind() != reflect.Struct {
		return jsonTypeNil
	}

	// Follow the path of attributes described by the key, so if the key was "foo.bar.baz" then look up
	// "foo" on the out value, then the "bar" attribute on that type, then the "baz" attribute on that type.
	// Once we exit the loop, 'actualType' should be the type of that nested "baz" field and we can
	// determine the correct JSON type from there.
	actualType := reflection.FlattenPointerType(outValue.Type())
	for i := 0; i < keyLength; i++ {
		field, ok := reflection.FindField(actualType, key[i])
		if !ok {
			return jsonTypeNil
		}
		actualType = reflection.FlattenPointerType(field.Type)
	}

	// Now that we have the Go type for the field that will ultimately be populated by this parameter/value,
	// we need to do a quick double check. The field's Go type might be some type alias for an int64 so the
	// natural choice for a JSON binding would be to use a number (which is what 't' will resolve to).
	//
	// But... what if the user provided the value "5m2s" for that field? If we blindly treat the value like
	// a number, we'll end up with JSON that looks like {"baz":5m2s} which is invalid. We need to quote that
	// value for it to remain valid JSON. So you only get to be a number/boolean if your parameter's value
	// looks like one of those values, too.
	//
	// The canonical use-case for this situation is if you define a custom type alias like this:
	//
	// type ISODuration int64
	//
	// You then implement the MarshalJSON() and UnmarshalJSON() functions so that it supports ISO duration formats
	// such as "PT3M49S". By looking at the Go type you'd think that the incoming param value should be a
	// JSON number (since the duration is an int64), but the value doesn't "look" like a number; it looks
	// like a freeform string. As a result, we need to build the binding JSON {"foo":"PT3M49S"} since we will
	// treat the right hand side as a string rather than {"foo":PT3M48S} which is not valid.
	t := b.typeToJSONType(actualType)
	if t == jsonTypeBool && !b.looksLikeBoolJSON(value) {
		return jsonTypeString
	}
	if t == jsonTypeNumber && !b.looksLikeNumberJSON(value) {
		return jsonTypeString
	}
	return t
}

// fieldToJSONType looks at the Go type of some field on a struct and returns the JSON data type
// that will most likely unmarshal to that field w/o an error.
func (b jsonBinder) typeToJSONType(actualType reflect.Type) jsonType {
	switch actualType.Kind() {
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

// looksLikeBoolJSON determines if the raw parameter value looks like a boolean value (i.e. true/false).
func (b jsonBinder) looksLikeBoolJSON(value string) bool {
	value = strings.ToLower(value)
	return value == "true" || value == "false"
}

// looksLikeNumberJSON determines if the raw parameter value looks like it can be formatted as a JSON
// number. Basically, does it only contain digits and a decimal point. Currently this only supports using
// periods as decimal points. A future iteration might support using /x/text/language to support commas
// as decimals points.
func (b jsonBinder) looksLikeNumberJSON(value string) bool {
	for _, r := range value {
		if r == '.' {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

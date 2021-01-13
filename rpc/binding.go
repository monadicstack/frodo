package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Binder performs the work of taking all meaningful values from an incoming request (body,
// path params, query string) and applying them to a Go struct (likely the "XxxRequest" for
// your service method).
type Binder interface {
	Bind(req *http.Request, params httprouter.Params, out interface{}) error
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
type jsonBinder struct {
}

func (b jsonBinder) Bind(req *http.Request, params httprouter.Params, out interface{}) error {
	if err := b.bindBody(req, params, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	if err := b.bindQueryString(req, params, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	if err := b.bindPathParams(req, params, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	return nil
}

func (b jsonBinder) bindBody(req *http.Request, _ httprouter.Params, out interface{}) error {
	if req.Body == nil {
		return nil
	}
	if req.Body == http.NoBody {
		return nil
	}
	return json.NewDecoder(req.Body).Decode(out)
}

func (b jsonBinder) bindQueryString(req *http.Request, _ httprouter.Params, out interface{}) error {
	jsonReader := b.queryStringToJSON(req)
	err := json.NewDecoder(jsonReader).Decode(out)
	if err != nil {
		return fmt.Errorf("bind query string: %w", err)
	}
	return nil
}

func (b jsonBinder) bindPathParams(_ *http.Request, params httprouter.Params, out interface{}) error {
	jsonReader := b.paramsToJSON(params)
	err := json.NewDecoder(jsonReader).Decode(out)
	if err != nil {
		return fmt.Errorf("bind path params: %w", err)
	}
	return nil
}

func (b jsonBinder) paramsToJSON(params httprouter.Params) io.Reader {
	paramJSON := &bytes.Buffer{}
	paramJSON.WriteString("{")
	for i, param := range params {
		if i > 0 {
			paramJSON.WriteString(", ")
		}
		b.writeAttributeJSON(paramJSON, param.Key, param.Value)
	}
	paramJSON.WriteString("}")
	return paramJSON
}

func (b jsonBinder) queryStringToJSON(req *http.Request) io.Reader {
	paramJSON := &bytes.Buffer{}
	paramJSON.WriteString("{")
	i := 0
	for key, values := range req.URL.Query() {
		if i > 0 {
			paramJSON.WriteString(", ")
		}
		b.writeAttributeJSON(paramJSON, key, values[0])
		i++
	}
	paramJSON.WriteString("}")
	return paramJSON
}

// writeAttributeJSON writes a key/value pair to a JSON object buffer. For instance, writing
// the key "foo" and the value "bar rules!" will result in this writing `"foo":"bar rules!"`
// to the buffer.
func (b jsonBinder) writeAttributeJSON(buf *bytes.Buffer, key string, value string) {
	buf.WriteString("\"")
	buf.WriteString(key)
	buf.WriteString("\":")
	switch {
	case looksLikeNumber(value):
		buf.WriteString(value)
	case looksLikeBool(value):
		buf.WriteString(value)
	case looksLikeStruct(value):
		buf.WriteString(value)
	default:
		buf.WriteString("\"")
		buf.WriteString(value)
		buf.WriteString("\"")
	}
}

func looksLikeNumber(value string) bool {
	if value == "" {
		return false
	}
	valueBytes := []byte(value)
	for _, b := range valueBytes {
		if b == '.' {
			continue
		}
		if b < '0' || b > '9' { // < '0' or > '9'
			return false
		}
	}
	return true
}

func looksLikeBool(value string) bool {
	if value == "" {
		return false
	}
	valueLower := strings.ToLower(value)
	return valueLower == "true" || valueLower == "false"
}

func looksLikeStruct(value string) bool {
	if value == "" {
		return false
	}
	return strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")
}

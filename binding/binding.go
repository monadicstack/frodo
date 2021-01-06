package binding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func Bind(req *http.Request, params httprouter.Params, out interface{}) error {
	if err := bindBody(req, params, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	if err := bindQueryString(req, params, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	if err := bindPathParams(req, params, out); err != nil {
		return fmt.Errorf("binding error: %w", err)
	}
	return nil
}

func bindBody(req *http.Request, _ httprouter.Params, out interface{}) error {
	if req.Body == nil {
		return nil
	}
	contentLength := strings.TrimSpace(req.Header.Get("Content-Type"))
	if contentLength == "" {
		return nil
	}
	if contentLength == "0" {
		return nil
	}
	return json.NewDecoder(req.Body).Decode(out)
}

func bindQueryString(req *http.Request, _ httprouter.Params, out interface{}) error {
	jsonReader := QueryStringToJSON(req)
	err := json.NewDecoder(jsonReader).Decode(out)
	if err != nil {
		return fmt.Errorf("bind query string: %w", err)
	}
	return nil
}

func bindPathParams(_ *http.Request, params httprouter.Params, out interface{}) error {
	jsonReader := ParamsToJSON(params)
	err := json.NewDecoder(jsonReader).Decode(out)
	if err != nil {
		return fmt.Errorf("bind path params: %w", err)
	}
	return nil
}

func ParamsToJSON(params httprouter.Params) io.Reader {
	paramJSON := &bytes.Buffer{}
	paramJSON.WriteString("{")
	for i, param := range params {
		if i > 0 {
			paramJSON.WriteString(", ")
		}
		writeAttributeJSON(paramJSON, param.Key, param.Value)
	}
	paramJSON.WriteString("}")
	return paramJSON
}

func QueryStringToJSON(req *http.Request) io.Reader {
	paramJSON := &bytes.Buffer{}
	paramJSON.WriteString("{")
	i := 0
	for key, values := range req.URL.Query() {
		if i > 0 {
			paramJSON.WriteString(", ")
		}
		writeAttributeJSON(paramJSON, key, values[0])
		i++
	}
	paramJSON.WriteString("}")
	return paramJSON
}

func writeAttributeJSON(buf *bytes.Buffer, key string, value string) {
	buf.WriteString("\"")
	buf.WriteString(key)
	buf.WriteString("\":")
	switch {
	case looksLikeNumber(value):
		buf.WriteString(value)
	case looksLikeBool(value):
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

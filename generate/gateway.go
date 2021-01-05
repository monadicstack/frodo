package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"text/template"

	"github.com/robsignorelli/expose/parser"
)

func Gateway(ctx *parser.Context, w io.Writer) error {
	buf := &bytes.Buffer{}
	err := gatewayTemplate.Execute(buf, ctx)
	if err != nil {
		return fmt.Errorf("error generating http gateway: %w", err)
	}

	sourceCode, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("[exposec] Unable to 'go fmt' gatway code: %v", err)
		_, err = w.Write(buf.Bytes())
	} else {
		_, err = w.Write(sourceCode)
	}
	if err != nil {
		return fmt.Errorf("error writing http gateway code: %w", err)
	}
	return nil
}

//--------------------------------

var gatewayTemplate = template.Must(template.New("gateway").Parse(`// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from {{.Path}}
// !!!!!!! DO NOT EDIT !!!!!!!
package {{ .Package }}

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/respond"
)

{{ $ctx := . }}
{{ range .Services }}
func New{{ .Name }}Gateway(service {{ .Name }}) *{{ .Name }}Gateway {
	gw := &{{.Name}}Gateway{
		service: service,
		router:  httprouter.New(),
	}

	{{ $service := . }}
	{{ range $service.Methods }}
	gw.router.{{ .GatewayMethod }}("/{{ .GatewayPath }}", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		reply := respond.To(w, req)

		request := {{ .Request.Name }}{}
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil {
			reply.Fail(err)
		}

		response, err := gw.service.{{ .Name }}(req.Context(), &request)
		reply.Ok(response, err)
	})
	{{ end }}

	return gw
}

type {{.Name}}Gateway struct {
	// The "real" implementation of the service that this gateway delegates to.
	service {{ .Name }}
	// The underlying HTTP router/mux that wraps all of our endpoints up into a single handler.
	router  *httprouter.Router
}

func (gw {{.Name}}Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.router.ServeHTTP(w, req)
}
{{end}}
`))

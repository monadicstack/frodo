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

func Server(ctx *parser.Context, w io.Writer) error {
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
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/respond"
	"github.com/robsignorelli/expose/gateway"
)

{{ $ctx := . }}
{{ range .Services }}
func New{{ .Name }}Gateway(service {{ .Name }}, options ...gateway.Option) *{{ .Name }}Gateway {
	gw := &{{.Name}}Gateway{
		HTTPGateway: gateway.New(options...),
		Service: service,
	}

	{{ $service := . }}
	{{ range $service.Methods }}
	gw.Router.{{ .HTTPMethod }}("{{ .HTTPPath }}", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := {{ .Request.Name }}{}
		if err := gw.Binder.Bind(req, params, &serviceRequest); err != nil {
			response.Fail(err)
			return
		}

		serviceResponse, err := gw.Service.{{ .Name }}(req.Context(), &serviceRequest)
		response.Reply({{ .HTTPStatus }}, serviceResponse, err)
	})
	{{ end }}

	return gw
}

type {{.Name}}Gateway struct {
	gateway.HTTPGateway
	// The "real" implementation of the service that this gateway delegates to.
	Service {{ .Name }}
}

func (gw {{ .Name }}Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.Middleware.ServeHTTP(w, req, gw.Router.ServeHTTP)
}
{{end}}
`))

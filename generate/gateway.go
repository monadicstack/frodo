package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"text/template"

	"github.com/robsignorelli/frodo/parser"
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
	"github.com/robsignorelli/frodo/rpc"
)

{{ $ctx := . }}
{{ range .Services }}
func New{{ .Name }}Gateway(service {{ .Name }}, options ...rpc.GatewayOption) rpc.Gateway {
	gw := rpc.NewGateway(options...)

	{{ $service := . }}
	{{ range $service.Methods }}
	gw.Router.{{ .HTTPMethod }}("{{ .HTTPPath }}", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := {{ .Request.Name }}{}
		if err := gw.Binder.Bind(req, params, &serviceRequest); err != nil {
			response.Fail(err)
			return
		}

		serviceResponse, err := service.{{ .Name }}(req.Context(), &serviceRequest)
		response.Reply({{ .HTTPStatus }}, serviceResponse, err)
	})
	{{ end }}

	return gw
}
{{end}}
`))

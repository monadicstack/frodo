package generate

import (
	"text/template"
)

// Once Go 1.16 comes out and we can embed files in the Go binary, I should pull this out
// into a separate template file and just embed that in the binary fs.
var TemplateGatewayGo = template.Must(template.New("gateway.go").Parse(`// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from {{ .Path }}
// !!!!!!! DO NOT EDIT !!!!!!!
package {{ .OutputPackage.Name }}

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/respond"
	"github.com/robsignorelli/frodo/rpc"
	"{{ .Package.Import }}"
)

{{ $ctx := . }}
{{ range .Services }}
func New{{ .Name }}Gateway(service {{ $ctx.Package.Name }}.{{ .Name }}, options ...rpc.GatewayOption) rpc.Gateway {
	gw := rpc.NewGateway(options...)

	{{ $service := . }}
	{{ range $service.Methods }}
	gw.Router.{{ .HTTPMethod }}(gw.PathPrefix + "{{ .HTTPPath }}", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		response := respond.To(w, req)

		serviceRequest := {{ $ctx.Package.Name }}.{{ .Request.Name }}{}
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

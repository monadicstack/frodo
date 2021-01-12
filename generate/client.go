package generate

import (
	"text/template"
)

// Once Go 1.16 comes out and we can embed files in the Go binary, I should pull this out
// into a separate template file and just embed that in the binary fs.
var TemplateClientGo = template.Must(template.New("client.go").Parse(`// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated client code from {{ .Path }}
// !!!!!!! DO NOT EDIT !!!!!!!
package {{ .OutputPackage.Name }}

import (
	"context"
	"fmt"

	"github.com/robsignorelli/frodo/rpc"
	"{{ .Package.Import }}"
)

{{ $ctx := . }}
{{ range .Services }}
// New{{ .Name }}Client creates an RPC client that conforms to the {{ .Name }} interface, but delegates
// work to remote instances. You must supply the base address of the remote service gateway instance or
// the load balancer for that service.
//
// You should be able to get a working service using default options (i.e. no options), but you can customize
// the HTTP client, define middleware, and more using client options. All of the ones that apply to the RPC
// client are named "WithClientXXX()".
func New{{ .Name }}Client(address string, options ...rpc.ClientOption) *{{ .Name }}Client {
	return &{{ .Name }}Client{
		Client: rpc.NewClient("{{ .Name }}", address, options...),
	}
}

// {{ .Name }}Client manages all interaction w/ a remote {{ .Name }} instance by letting you invoke functions
// on this instance as if you were doing it locally (hence... RPC client). You shouldn't instantiate this
// manually. Instead, you should utilize the New{{ .Name }}Client() function to properly set this up.
type {{ .Name }}Client struct {
	rpc.Client
}

{{ $service := . }}
{{ range .Methods }}
func (client *{{ $service.Name }}Client) {{ .Name }} (ctx context.Context, request *{{ $ctx.Package.Name }}.{{ .Request.Name }}) (*{{ $ctx.Package.Name }}.{{ .Response.Name }}, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &{{ $ctx.Package.Name }}.{{ .Response.Name }}{}
	err := client.Invoke(ctx, "{{ .HTTPMethod }}", "{{ .HTTPPath }}", request, response)
	return response, err
}
{{ end }}

// {{ .Name }}Proxy fully implements the {{ .Name }} interface, but delegates all operations to a "real"
// instance of the service. You can embed this type in a struct of your choice so you can "override" or
// decorate operations as you see fit. Any operations on {{ .Name }} that you don't explicitly define will
// simply delegate to the default implementation of the underlying 'Service' value.
//
// Since you have access to the underlying service, you are able to both implement custom handling logic AND
// call the "real" implementation, so this can be used as special middleware that applies to only certain operations.
type {{ .Name }}Proxy struct {
	Service {{ $ctx.Package.Name }}.{{ .Name }}
}

{{ range .Methods }}
func (proxy *{{ $service.Name }}Proxy) {{ .Name }} (ctx context.Context, request *{{ $ctx.Package.Name }}.{{ .Request.Name }}) (*{{ $ctx.Package.Name }}.{{ .Response.Name }}, error) {
	return proxy.Service.{{ .Name }}(ctx, request)
}
{{ end }}
{{ end }}
`))

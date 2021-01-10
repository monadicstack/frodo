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

func Client(ctx *parser.Context, w io.Writer) error {
	buf := &bytes.Buffer{}
	err := clientTemplate.Execute(buf, ctx)
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

var clientTemplate = template.Must(template.New("gateway").Parse(`// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated client code from {{ .Path }}
// !!!!!!! DO NOT EDIT !!!!!!!
package {{ .Package }}

import (
	"context"
	"fmt"

	"github.com/robsignorelli/frodo/rpc"
)

{{ $ctx := . }}
{{ range .Services }}
func New{{ .Name }}Client(address string, options ...rpc.ClientOption) *{{.Name}}Client {
	fmt.Println(">>>> Creating client")
	return &{{.Name}}Client{
		Client: rpc.NewClient("{{ .Name }}", address, options...),
	}
}

type {{ .Name }}Client struct {
	rpc.Client
}

{{ $service := . }}
{{ range .Methods }}
func (client *{{ $service.Name }}Client) {{ .Name }} (ctx context.Context, request *{{ .Request.Name }}) (*{{ .Response.Name }}, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &{{ .Response.Name }}{}
	err := client.Invoke(ctx, "{{ .HTTPMethod }}", "{{ .HTTPPath }}", request, response)
	return response, err
}
{{ end }}

type {{ .Name }}Proxy struct {
	Service {{ .Name }}
}

{{ range .Methods }}
func (proxy *{{ $service.Name }}Proxy) {{ .Name }} (ctx context.Context, request *{{ .Request.Name }}) (*{{ .Response.Name }}, error) {
	return proxy.Service.{{ .Name }}(ctx, request)
}
{{ end }}
{{ end }}

`))

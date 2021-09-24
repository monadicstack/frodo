package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/monadicstack/frodo/generate"
	"github.com/monadicstack/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateClientRequest contains all of the CLI options used in the "frodo client" command.
type GenerateClientRequest struct {
	templateOption
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
	// Language is the programming language for the client to generate (the "--language" option)
	Language string
}

// GenerateClient handles the registration and execution of the 'frodo client' CLI subcommand.
type GenerateClient struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateClient) Command() *cobra.Command {
	request := &GenerateClientRequest{}
	cmd := &cobra.Command{
		Use:   "client [flags] FILENAME",
		Short: "Process a Go source file with your service interface to generate an RPC client proxy for your service(s).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			request.InputFileName = args[0]
			crapPants(c.Exec(request))
		},
	}
	cmd.Flags().StringVar(&request.Language, "language", "go", "The file extension of the target language (e.g. 'go' or 'js')")
	cmd.Flags().StringVar(&request.Template, "template", "", "Path to a custom Go template file used to generate this artifact.")
	return cmd
}

// Exec takes all of the parsed CLI flags and generates the target client artifact.
func (c GenerateClient) Exec(request *GenerateClientRequest) error {
	switch strings.ToLower(request.Language) {
	case "go", "":
		return c.generate(request, request.ToFileTemplate("client.go"))
	case "js", "javascript", "node", "nodejs":
		return c.generate(request, request.ToFileTemplate("client.js"))
	case "java":
		return c.generate(request, request.ToFileTemplate("client.java"))
	case "dart", "flutter":
		return c.generate(request, request.ToFileTemplate("client.dart"))
	default:
		return fmt.Errorf("unsupported client language")
	}
}

// generate parses the input service definition file and creates an output client/gateway
// code, writing it to the output gen/ directory.
func (c GenerateClient) generate(request *GenerateClientRequest, artifact generate.FileTemplate) error {
	log.Printf("Parsing service definition: %s", request.InputFileName)
	ctx, err := parser.ParseFile(request.InputFileName)
	if err != nil {
		return err
	}

	log.Printf("Generating '%s'", artifact.Name)
	return generate.File(ctx, artifact)
}

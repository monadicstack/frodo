package cli

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/robsignorelli/frodo/generate"
	"github.com/robsignorelli/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateClientRequest contains all of the CLI options used in the "frodo client" command.
type GenerateClientRequest struct {
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
		Use:  "client",
		Args: cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.Exec(request)
		},
	}
	cmd.Flags().StringVar(&request.InputFileName, "input", "", "Path to the Go file w/ your service interface.")
	cmd.Flags().StringVar(&request.Language, "language", "go", "The file extension for the language to output (e.g. 'go' or 'js')")

	_ = cmd.MarkFlagRequired("input")
	return cmd
}

// Exec takes all of the parsed CLI flags and generates the target client artifact.
func (c GenerateClient) Exec(request *GenerateClientRequest) error {
	switch strings.ToLower(request.Language) {
	case "go", "":
		return c.generate(request, generate.TemplateClientGo)
	case "js":
		return c.generate(request, generate.TemplateClientJS)
	default:
		return fmt.Errorf("unsupported client language")
	}
}

// generate parses the input service definition file and creates an output client/gateway
// code, writing it to the output gen/ directory.
func (c GenerateClient) generate(request *GenerateClientRequest, artifactTemplate *template.Template) error {
	log.Printf("[frodo] Parsing service definitions: %s", request.InputFileName)
	ctx, err := parser.ParseFile(request.InputFileName)
	if err != nil {
		return err
	}

	log.Printf("[frodo] Generating artifact '%s'", artifactTemplate.Name())
	return generate.Artifact(ctx, request.InputFileName, artifactTemplate)
}

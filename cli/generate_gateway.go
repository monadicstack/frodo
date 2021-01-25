package cli

import (
	"log"

	"github.com/robsignorelli/frodo/generate"
	"github.com/robsignorelli/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateGatewayRequest contains all of the CLI options used in the "frodo client" command.
type GenerateGatewayRequest struct {
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
}

// GenerateGateway handles the registration and execution of the 'frodo gateway' CLI subcommand.
type GenerateGateway struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateGateway) Command() *cobra.Command {
	request := &GenerateGatewayRequest{}
	cmd := &cobra.Command{
		Use:  "gateway",
		Args: cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.Exec(request)
		},
	}
	cmd.Flags().StringVar(&request.InputFileName, "input", "", "Path to the Go file w/ your service interface.")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

// Exec actually executes the parsing/generating logic creating the gateway for the given declaration.
func (c GenerateGateway) Exec(request *GenerateGatewayRequest) error {
	log.Printf("[frodo] Parsing service definitions: %s", request.InputFileName)
	ctx, err := parser.ParseFile(request.InputFileName)
	if err != nil {
		return err
	}

	log.Printf("[frodo] Generating artifact '%s'", generate.TemplateGatewayGo.Name())
	return generate.Artifact(ctx, request.InputFileName, generate.TemplateGatewayGo)
}

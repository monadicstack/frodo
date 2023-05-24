package cli

import (
	"log"

	"github.com/davidrenne/frodo/generate"
	"github.com/davidrenne/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateMockRequest contains all of the CLI options used in the "frodo mock" command.
type GenerateMockRequest struct {
	templateOption
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
}

// GenerateMock handles the registration and execution of the 'frodo mock' CLI subcommand.
type GenerateMock struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateMock) Command() *cobra.Command {
	request := &GenerateMockRequest{}
	cmd := &cobra.Command{
		Use:   "mock [flags] FILENAME",
		Short: "Creates a mock instance of your service for unit testing.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			request.InputFileName = args[0]
			crapPants(c.Exec(request))
		},
	}
	cmd.Flags().StringVar(&request.Template, "template", "", "Path to a custom Go template file used to generate this artifact.")
	return cmd
}

// Exec takes all of the parsed CLI flags and generates the target mock service artifact.
func (c GenerateMock) Exec(request *GenerateMockRequest) error {
	log.Printf("Parsing service definitions: %s", request.InputFileName)
	ctx, err := parser.ParseFile(request.InputFileName)
	if err != nil {
		return err
	}

	artifact := request.ToFileTemplate("mock.go")
	log.Printf("Generating artifact '%s'", artifact.Name)
	return generate.File(ctx, artifact)
}

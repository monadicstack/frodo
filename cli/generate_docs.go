package cli

import (
	"log"

	"github.com/davidrenne/frodo/generate"
	"github.com/davidrenne/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateDocsRequest contains all of the CLI options used in the "frodo docs" command.
type GenerateDocsRequest struct {
	templateOption
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
}

// GenerateDocs handles the registration and execution of the 'frodo docs' CLI subcommand.
type GenerateDocs struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateDocs) Command() *cobra.Command {
	request := &GenerateDocsRequest{}
	cmd := &cobra.Command{
		Use:   "docs [flags] FILENAME",
		Short: "Generates the API documentation for your service that can be distributed to users.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			request.InputFileName = args[0]
			crapPants(c.Exec(request))
		},
	}
	cmd.Flags().StringVar(&request.Template, "template", "", "Path to a custom Go template file used to generate this artifact.")
	return cmd
}

// Exec takes all of the parsed CLI flags and generates the service's documentation artifact(s).
func (c GenerateDocs) Exec(request *GenerateDocsRequest) error {
	log.Printf("Parsing service definitions: %s", request.InputFileName)
	ctx, err := parser.ParseFile(request.InputFileName)
	if err != nil {
		return err
	}

	artifact := request.ToFileTemplate("openapi.yml")
	log.Printf("Generating artifact '%s'", artifact.Name)
	return generate.File(ctx, artifact)
}

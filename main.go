package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/robsignorelli/frodo/generate"
	"github.com/robsignorelli/frodo/parser"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "frodo",
		Short: "A code generator for Go-based (micro)services that creates RPC clients/gateways.",
	}
	rootCmd.AddCommand(GatewayCommand{}.Cobra())
	rootCmd.AddCommand(ClientCommand{}.Cobra())
	rootCmd.AddCommand(CreateCommand{}.Cobra())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GatewayCommand handles the registration and execution of the 'frodo gateway' CLI subcommand.
type GatewayCommand struct{}

// Cobra creates the Cobra struct describing this CLI command and its options.
func (c GatewayCommand) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "gateway",
		Args: cobra.MaximumNArgs(0),
		RunE: c.Run,
	}
	cmd.Flags().StringP("input", "i", "", "Path to the Go file w/ your service interface.")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

// Run handles when the user actually executes 'frodo gateway'; generating a Go-based gateway
// type that can be fed to http.ListenAndServe() in order to expose a service over RPC.
func (c GatewayCommand) Run(cmd *cobra.Command, _ []string) error {
	inputFileName := cmd.Flag("input").Value.String()
	return generateArtifact(inputFileName, generate.TemplateGatewayGo)
}

// ClientCommand handles the registration and execution of the 'frodo client' CLI subcommand.
type ClientCommand struct{}

// Cobra creates the Cobra struct describing this CLI command and its options.
func (c ClientCommand) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "client",
		Args: cobra.MaximumNArgs(0),
		RunE: c.Run,
	}
	cmd.Flags().StringP("input", "i", "", "Path to the Go file w/ your service interface.")
	_ = cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("language", "l", "go", "The file extension for the language to output (e.g. 'go' or 'js')")
	return cmd
}

// Run handles when the user executes 'frodo client'; generating a strongly typed RPC client that
// accepts/returns data structures, but handles all of the HTTP/RPC magic under the hood.
func (c ClientCommand) Run(cmd *cobra.Command, _ []string) error {
	inputFileName := cmd.Flag("input").Value.String()
	lang := cmd.Flag("language").Value.String()
	switch strings.ToLower(lang) {
	case "go":
		return generateArtifact(inputFileName, generate.TemplateClientGo)
	case "js":
		return generateArtifact(inputFileName, generate.TemplateClientJS)
	default:
		return fmt.Errorf("unsupported client language")
	}
}

// generateArtifact parses the input service definition file and creates an output client/gateway
// code, writing it to the output gen/ directory.
func generateArtifact(inputFileName string, artifactTemplate *template.Template) error {
	log.Printf("[frodo] Parsing service definitions: %s", inputFileName)
	ctx, err := parser.ParseFile(inputFileName)
	if err != nil {
		return err
	}

	log.Printf("[frodo] Generating artifact '%s'", artifactTemplate.Name())
	err = generate.Artifact(ctx, inputFileName, artifactTemplate)
	if err != nil {
		return err
	}

	return nil
}

// CreateCommand is the scaffolding command that creates a new service directory and a minimal
// template Go declaration file w/ your service and some sample operations.
type CreateCommand struct{}

// Cobra creates the Cobra struct describing this CLI command and its options.
func (c CreateCommand) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create",
		Args: cobra.MaximumNArgs(0),
		RunE: c.Run,
	}
	cmd.Flags().StringP("service", "s", "", "The name of the service to create (doesn't have to end in 'Service')")
	cmd.Flags().StringP("dir", "d", "", "Path to the directory where we'll write the Go file (defaults to new directory named after the service)")
	cmd.Flags().BoolP("force", "f", false, "Overwrite declaration/handler source code files if they exist. (default=false)")

	_ = cmd.MarkFlagRequired("service")
	return cmd
}

// Run handles when the user executes 'frodo create'; scaffolding a new service directory/declaration.
func (c CreateCommand) Run(cmd *cobra.Command, _ []string) error {
	return generate.ServiceScaffold(generate.ServiceScaffoldRequest{
		ServiceName: cmd.Flag("service").Value.String(),
		Directory:   cmd.Flag("dir").Value.String(),
		Force:       cmd.Flag("force").Value.String() == "true",
	})
}

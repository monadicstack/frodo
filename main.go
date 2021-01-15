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
		Use:   "frodoc",
		Short: "A code generator for Go-based (micro)services that creates RPC clients/gateways.",
	}
	rootCmd.AddCommand(GatewayCommand{}.Cobra())
	rootCmd.AddCommand(ClientCommand{}.Cobra())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GatewayCommand handles the registration and execution of the 'frodoc gateway' CLI subcommand.
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

// Run handles when the user actually executes 'frodoc gateway'; generating a Go-based gateway
// type that can be fed to http.ListenAndServe() in order to expose a service over RPC.
func (c GatewayCommand) Run(cmd *cobra.Command, _ []string) error {
	inputFileName := cmd.Flags().Lookup("input").Value.String()
	return generateArtifact(inputFileName, generate.TemplateGatewayGo)
}

// ClientCommand handles the registration and execution of the 'frodoc client' CLI subcommand.
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

	cmd.Flags().StringP("language", "l", "", "The file extension for the language to output (e.g. 'go' or 'js')")
	_ = cmd.MarkFlagRequired("language")
	return cmd
}

// Run handles when the user executes 'frodoc client'; generating a strongly typed RPC client that
// accepts/returns data structures, but handles all of the HTTP/RPC magic under the hood.
func (c ClientCommand) Run(cmd *cobra.Command, _ []string) error {
	inputFileName := cmd.Flags().Lookup("input").Value.String()
	lang := cmd.Flags().Lookup("language").Value.String()
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
	log.Printf("[frodoc] Parsing service definitions: %s", inputFileName)
	ctx, err := parser.ParseFile(inputFileName)
	if err != nil {
		return err
	}

	log.Printf("[frodoc] Generating artifact '%s'", artifactTemplate.Name())
	err = generate.Artifact(ctx, inputFileName, artifactTemplate)
	if err != nil {
		return err
	}

	return nil
}

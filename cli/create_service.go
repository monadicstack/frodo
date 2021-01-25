package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/robsignorelli/frodo/parser"
	"github.com/spf13/cobra"
)

// CreateServiceRequest contains the inputs from our "frodo create" CLI command.
type CreateServiceRequest struct {
	// ServiceName is the value of the --service argument.
	ServiceName string
	// Directory is the value of the --dir argument which defines where the new .go files will be written.
	Directory string
	// Force is the status of the --force flag to overwrite files if they already exist.
	Force bool
}

// CreateService is the scaffolding command that creates a new service directory and a minimal
// template Go declaration file w/ your service and some sample operations.
type CreateService struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c CreateService) Command() *cobra.Command {
	args := &CreateServiceRequest{}
	cmd := &cobra.Command{
		Use:  "create",
		Args: cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return c.Exec(args)
		},
	}
	cmd.Flags().StringVar(&args.ServiceName, "service", "", "The name of the service to create (doesn't have to end in 'Service')")
	cmd.Flags().StringVar(&args.Directory, "dir", "", "Path to the directory where we'll write the Go file (defaults to new directory named after the service)")
	cmd.Flags().BoolVar(&args.Force, "force", false, "Overwrite declaration/handler source code files if they exist. (default=false)")

	_ = cmd.MarkFlagRequired("service")
	return cmd
}

// CreateService creates the bare minimum code required to have a frodo-powered service. It
// creates a directory for the service code to live, a declaration file that contains the
// interface and model definitions, and a skeleton implementation. These all help establish some
// of the base patterns you should use when working with frodo services.
func (c CreateService) Exec(request *CreateServiceRequest) error {
	shortName := request.ServiceName
	shortName = strings.TrimSuffix(shortName, "Service")
	shortName = strings.TrimSuffix(shortName, "service")

	shortNameLower := strings.ToLower(shortName)
	shortNameTitle := strings.Title(shortName)

	ctx := createServiceContext{
		Request:        request,
		ShortName:      shortNameTitle,
		ShortNameLower: shortNameLower,
		ServiceName:    shortNameTitle + "Service",
		HandlerName:    shortNameTitle + "ServiceHandler",
		Directory:      request.Directory,
	}

	if ctx.Directory == "" {
		ctx.Directory = ctx.ShortNameLower
	}
	ctx.Package = filepath.Base(ctx.Directory)
	ctx.Paths.Service = filepath.Join(ctx.Directory, ctx.ShortNameLower+"_service.go")
	ctx.Paths.Handler = filepath.Join(ctx.Directory, ctx.ShortNameLower+"_handler.go")
	ctx.Paths.Makefile = filepath.Join(ctx.Directory, "makefile")
	ctx.Paths.Main = filepath.Join(ctx.Directory, "cmd", "main.go")

	// Create the service declaration and handler stubs.
	if err := os.MkdirAll(ctx.Directory, 0777); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(ctx.Directory, "cmd"), 0777); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, createServiceTemplate, ctx.Paths.Service); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, createHandlerTemplate, ctx.Paths.Handler); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, createMakefileTemplate, ctx.Paths.Makefile); err != nil {
		return err
	}

	info, err := parser.ParseFile(ctx.Paths.Service)
	if err != nil {
		return err
	}
	ctx.PackageImport = info.Package.Import
	if err := scaffoldTemplate(ctx, createMainTemplate, ctx.Paths.Main); err != nil {
		return err
	}

	// Run those stubs through 'frodo' to create the gateway and client generated artifacts.
	generateClientRequest := &GenerateClientRequest{InputFileName: ctx.Paths.Service}
	if err := (GenerateClient{}).Exec(generateClientRequest); err != nil {
		return err
	}
	generateGatewayRequest := &GenerateGatewayRequest{InputFileName: ctx.Paths.Service}
	if err := (GenerateGateway{}).Exec(generateGatewayRequest); err != nil {
		return err
	}
	return nil
}

func scaffoldTemplate(ctx createServiceContext, t *template.Template, path string) error {
	// Only allow you to overwrite the file if you included the --force argument.
	_, err := os.Stat(path)
	if !os.IsNotExist(err) && !ctx.Request.Force {
		return fmt.Errorf("unable to open %s: already exists (use --force to overwrite it)", path)
	}

	_ = os.Remove(path)
	outputFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", path, err)
	}
	defer outputFile.Close()

	err = t.Execute(outputFile, ctx)
	if err != nil {
		return fmt.Errorf("unable to eval code template: %s: %w", path, err)
	}
	return nil
}

type createServiceContext struct {
	// Request are the raw incoming params to the scaffolding operation we're processing.
	Request *CreateServiceRequest
	// ShortNameLower is the name of the service w/o the "Service" suffix (e.g. "Greeter").
	ShortName string
	// ShortNameLower is the name of the service w/o the "Service" suffix in all lower case (e.g. "greeter").
	ShortNameLower string
	// ServiceName is the "cleaned up" version used for the service interface (e.g. "GreeterService").
	ServiceName string
	// HandlerName is the name of the struct for the real implementation (e.g. "GreeterServiceHandler").
	HandlerName string
	// Directory is the directory where we will put the declaration file for the service.
	Directory string
	// Package is the name of the package that the new service will belong to.
	Package string
	// PackageImport is the full import path for this service within the module.
	PackageImport string
	// Paths contains the directory/filename paths to the various assets we're creating.
	Paths struct {
		Service  string
		Handler  string
		Makefile string
		Main     string
	}
}

var createServiceTemplate = template.Must(template.New("service.go").Parse(`package {{ .Package }}

import (
	"context"
)

// {{ .ServiceName }} is a service that...
type {{ .ServiceName }} interface  {
    // Create saves a new {{ .ShortName }} record to the database. 
	Create(context.Context, *CreateRequest) (*CreateResponse, error)
}

type CreateRequest struct {
	Name        string
	Description string
}

type CreateResponse struct {
	ID          string
	Name        string
	Description string
}
`))

var createHandlerTemplate = template.Must(template.New("service_handler.go").Parse(`package {{ .Package }}

import (
	"context"

	"github.com/robsignorelli/frodo/rpc/errors"
)

// {{ .HandlerName }} implements all of the "real" functionality for the {{ .ServiceName }}.
type {{ .ServiceName }}Handler struct{}

func (svc *{{ .HandlerName }}) Create(ctx context.Context, request *CreateRequest) (*CreateResponse, error) {
	if request.Name == "" {
		return nil, errors.BadRequest("create: name is required")
	}
	// Beep boop beep... pretend we're doing real work...
	return &CreateResponse{
		ID:          "1234",
		Name:        request.Name,
		Description: request.Description,
	}, nil
}
`))

var createMakefileTemplate = template.Must(template.New("makefile").Parse(`
run: build
	out/{{ .ShortNameLower }}

frodo:
	frodo gateway --input={{ .ShortNameLower }}_service.go && \
	frodo client  --input={{ .ShortNameLower }}_service.go

build: frodo
	go build -o out/{{ .ShortNameLower }} cmd/main.go

test:
	go test ./...
`))

var createMainTemplate = template.Must(template.New("main").Parse(`package main

import (
	"net/http"

	"{{ .PackageImport }}"
	"{{ .PackageImport }}/gen"
)

func main() {
	serviceHandler := {{ .Package }}.{{ .HandlerName }}{}
	gateway := {{ .Package }}rpc.New{{ .ServiceName }}Gateway(&serviceHandler)
	http.ListenAndServe(":9001", gateway)
}
`))

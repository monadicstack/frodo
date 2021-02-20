package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/monadicstack/frodo/parser"
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
	// Port defines which HTTP port you want the RPC/HTTP gateway to run on by default.
	Port int
}

// CreateService is the scaffolding command that creates a new service directory and a minimal
// template Go declaration file w/ your service and some sample operations.
type CreateService struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c CreateService) Command() *cobra.Command {
	request := &CreateServiceRequest{}
	cmd := &cobra.Command{
		Use:   "create [flags] SERVICE_NAME",
		Short: "Creates a new service in your project with all of the code required to run.",
		Long:  "This creates a new package in your project for the service. It creates 4 different Go files: your service declaration (interface/structs), your service handler/implementation, the frodo RPC client, and the frodo RPC/API gateway. This also creates a makefile with convenience targets that regenerate your frodo RPC artifacts, build, test, and run your new service.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			request.ServiceName = args[0]
			return c.Exec(request)
		},
	}
	cmd.Flags().StringVar(&request.Directory, "dir", "", "Path to the directory where we'll write the Go file (defaults to new directory named after the service)")
	cmd.Flags().BoolVar(&request.Force, "force", false, "Overwrite declaration/handler source code files if they exist.")
	cmd.Flags().IntVar(&request.Port, "port", 0, "When generating main(), what port will the RPC/API gateway run on? (default = random port between 9000-9999)")
	return cmd
}

// Exec creates the bare minimum code required to have a frodo-powered service. It
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
		Port:           request.Port,
	}

	// Let the user pick their port, but if they didn't, just assign a random one between 9000 and 9999
	if ctx.Port == 0 {
		rand.Seed(time.Now().UnixNano())
		minPort := 9000
		maxPort := 9999
		ctx.Port = rand.Intn(maxPort-minPort+1) + minPort
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
	ctx.PackageImport = info.InputPackage.Import
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
	// Port is the HTTP port we will have main() listen on to expose the RPC gateway.
	Port int
	// Paths contains the directory/filename paths to the various assets we're creating.
	Paths struct {
		Service  string
		Handler  string
		Makefile string
		Main     string
	}
}

var createServiceTemplate = template.Must(template.New("service.go").Parse(`package {{.InputPackage }}

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

var createHandlerTemplate = template.Must(template.New("service_handler.go").Parse(`package {{.InputPackage }}

import (
	"context"

	"github.com/monadicstack/frodo/rpc/errors"
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

var createMakefileTemplate = template.Must(template.New("makefile").Parse(`#
# Local development only. This builds and executes the service in a local process.
#
run: build
	out/{{ .ShortNameLower }}

#
# Runs {{ .ShortNameLower }}_service.go through the 'frodo' code generator to spit out
# the latest and greatest RPC client/gateway code.
#
frodo:
	frodo gateway {{ .ShortNameLower }}_service.go && \
	frodo client  {{ .ShortNameLower }}_service.go

#
# If you add Frodo-based "//go:generate" comments to your service, generate your Frodo
# artifacts using that method instead.
#
generate:
	go generate .

#
# Rebuilds the binary for this service. We will "re-frodo" the service declaration beforehand
# so that any modifications to your service are always reflected in your client/gateway code
# without you having to think about it.
#
build: frodo
	go build -o out/{{ .ShortNameLower }} cmd/main.go

#
# This target hacks the Gibson; what do you think 'test' does? It runs through all of
# the test suites for this service.
#
test:
	go test ./...
`))

var createMainTemplate = template.Must(template.New("main").Parse(`package main

import (
	"net/http"

	"{{.PackageImport }}"
	"{{.PackageImport }}/gen"
)

func main() {
	serviceHandler := {{.InputPackage }}.{{ .HandlerName }}{}
	gateway := {{.InputPackage }}rpc.New{{ .ServiceName }}Gateway(&serviceHandler)
	http.ListenAndServe(":{{ .Port }}", gateway)
}
`))

package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monadicstack/frodo/generate"
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

	if err := scaffoldTemplate(ctx, "templates/create/service.go.tmpl", ctx.Paths.Service); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, "templates/create/service_handler.go.tmpl", ctx.Paths.Handler); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, "templates/create/makefile.tmpl", ctx.Paths.Makefile); err != nil {
		return err
	}

	info, err := parser.ParseFile(ctx.Paths.Service)
	if err != nil {
		return err
	}
	ctx.PackageImport = info.InputPackage.Import
	if err := scaffoldTemplate(ctx, "templates/create/main.go.tmpl", ctx.Paths.Main); err != nil {
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

func scaffoldTemplate(ctx createServiceContext, templatePath string, path string) error {
	t := generate.NewStandardTemplate(templatePath, templatePath)

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

	code, err := t.Eval(ctx)
	if err != nil {
		return fmt.Errorf("unable to eval code template: %s: %w", path, err)
	}

	_, err = outputFile.Write(code)
	if err != nil {
		return fmt.Errorf("unable to write code artifact: %s: %w", path, err)
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

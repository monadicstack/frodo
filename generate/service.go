package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ServiceScaffold creates the bare minimum code required to have a frodo-powered service. It
// creates a directory for the service code to live, a declaration file that contains the
// interface and model definitions, and a skeleton implementation. These all help establish some
// of the base patterns you should use when working with frodo services.
func ServiceScaffold(request ServiceScaffoldRequest) error {
	shortName := request.ServiceName
	shortName = strings.TrimSuffix(shortName, "Service")
	shortName = strings.TrimSuffix(shortName, "service")

	shortNameLower := strings.ToLower(shortName)
	shortNameTitle := strings.Title(shortName)

	ctx := scaffoldServiceContext{
		Request:        request,
		ShortName:      shortNameTitle,
		ShortNameLower: shortNameLower,
		InterfaceName:  shortNameTitle + "Service",
		HandlerName:    shortNameTitle + "ServiceHandler",
		Directory:      request.Directory,
	}

	if ctx.Directory == "" {
		ctx.Directory = ctx.ShortNameLower
	}
	ctx.Package = filepath.Base(ctx.Directory)

	if err := scaffoldDirectory(ctx); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, serviceDeclarationTemplate); err != nil {
		return err
	}
	if err := scaffoldTemplate(ctx, serviceHandlerTemplate); err != nil {
		return err
	}
	return nil
}

// ServiceScaffoldRequest contains the inputs from our "frodo create" CLI command.
type ServiceScaffoldRequest struct {
	// ServiceName is the value of the --service argument.
	ServiceName string
	// Directory is the value of the --dir argument which defines where the new .go files will be written.
	Directory string
	// Force is the status of the --force flag to overwrite files if they already exist.
	Force bool
}

func scaffoldDirectory(ctx scaffoldServiceContext) error {
	info, err := os.Stat(ctx.Directory)
	if os.IsNotExist(err) {
		return os.MkdirAll(ctx.Directory, 0777)
	}
	if !info.IsDir() {
		return fmt.Errorf("unable to create service: '%s' is not a directory", ctx.Directory)
	}
	return nil
}

func scaffoldTemplate(ctx scaffoldServiceContext, t *template.Template) error {
	path := filepath.Join(ctx.Directory, ctx.ShortNameLower+"_"+t.Name())

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

type scaffoldServiceContext struct {
	// Request are the raw incoming params to the scaffolding operation we're processing.
	Request ServiceScaffoldRequest
	// ShortNameLower is the name of the service w/o the "Service" suffix (e.g. "Greeter").
	ShortName string
	// ShortNameLower is the name of the service w/o the "Service" suffix in all lower case (e.g. "greeter").
	ShortNameLower string
	// InterfaceName is the "cleaned up" version used for the service interface (e.g. "GreeterService").
	InterfaceName string
	// HandlerName is the name of the struct for the real implementation (e.g. "GreeterServiceHandler").
	HandlerName string
	// Directory is the directory where we will put the declaration file for the service.
	Directory string
	// Package is the name of the package that the new service will belong to.
	Package string
}

var serviceDeclarationTemplate = parseArtifactTemplate("service.go", `package {{ .Package }}

import (
	"context"
)

// {{ .InterfaceName }} is a service that...
type {{ .InterfaceName }} interface  {
    // Lookup fetches a {{ .ShortName }} record by its unique identifier. 
	Lookup(context.Context, *LookupRequest) (*LookupResponse, error)
}

type LookupRequest struct {
	ID string
}

type LookupResponse struct {
	ID   string
	Name string
}
`)

var serviceHandlerTemplate = parseArtifactTemplate("service_handler.go", `package {{ .Package }}

import (
	"context"

	"github.com/monadicstack/frodo/rpc/errors"
)

// {{ .HandlerName }} implements all of the "real" functionality for the {{ .InterfaceName }}.
type {{ .InterfaceName }}Handler struct{}

func (svc *{{ .HandlerName }}) Lookup(ctx context.Context, request *LookupRequest) (*LookupResponse, error) {
	if request.ID == "" {
		return nil, errors.BadRequest("lookup: id is required")
	}
	return &LookupResponse{ID: request.ID, Name: "Beetlejuice"}, nil
}
`)

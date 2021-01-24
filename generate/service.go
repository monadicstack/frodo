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
func ServiceScaffold(serviceName string, path string) error {
	shortName := serviceName
	shortName = strings.TrimSuffix(shortName, "Service")
	shortName = strings.TrimSuffix(shortName, "service")

	shortNameLower := strings.ToLower(shortName)
	shortNameTitle := strings.Title(shortName)

	ctx := scaffoldServiceContext{
		RawName:        serviceName,
		ShortName:      shortNameTitle,
		ShortNameLower: shortNameLower,
		InterfaceName:  shortNameTitle + "Service",
		HandlerName:    shortNameTitle + "ServiceHandler",
		Path:           path,
	}

	if path == "" {
		ctx.Path = ctx.ShortNameLower
	}
	ctx.Package = filepath.Base(ctx.Path)

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

func scaffoldDirectory(ctx scaffoldServiceContext) error {
	dirInfo, err := os.Stat(ctx.Path)
	if os.IsNotExist(err) {
		return os.MkdirAll(ctx.Path, 0777)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("unable to create service: '%s' is not a directory", ctx.Path)
	}
	return nil
}

func scaffoldTemplate(ctx scaffoldServiceContext, t *template.Template) error {
	path := filepath.Join(ctx.Path, ctx.ShortNameLower+t.Name())

	_ = os.Remove(path)
	outputFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to open file: %s: %w", path, err)
	}
	defer outputFile.Close()

	err = t.Execute(outputFile, ctx)
	if err != nil {
		return fmt.Errorf("unable to eval code template: %s: %w", path, err)
	}
	return nil
}

type scaffoldServiceContext struct {
	// RawName is the exact input that the user provided when invoking the scaffolding command.
	RawName string
	// ShortNameLower is the name of the service w/o the "Service" suffix (e.g. "Greeter")
	ShortName string
	// ShortNameLower is the name of the service w/o the "Service" suffix in all lower case (e.g. "greeter")
	ShortNameLower string
	// InterfaceName is the "cleaned up" version used for the service interface (e.g. "GreeterService")
	InterfaceName string
	// HandlerName is the name of the struct for the real implementation (e.g. "GreeterServiceHandler")
	HandlerName string
	// Path is the directory where we will put the declaration file for the service.
	Path string
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

	"github.com/robsignorelli/frodo/rpc/errors"
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

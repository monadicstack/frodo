package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"net/http"
	"path/filepath"
	"strings"
)

// Context wrangles all of the captured data about your input service declaration file. It tracks
// the module/package information, the service(s) that were defined, the request/response structs
// that were defined in the file, etc. It's the output of Parse() and is the input value when we
// evaluate Go templates to generate other source files based on this service definition info.
type Context struct {
	// FileSet is the collection of related files we're going to give to the Go AST parser.
	FileSet *token.FileSet
	// File is the entire syntax tree from when we parsed your input file.
	File *ast.File
	// Path is the relative path to the service definition file we're parsing.
	Path string
	// AbsolutePath is the absolute path to the service definition file we're parsing.
	AbsolutePath string
	// Package contains information about the package where the service definition resides.
	Package *PackageDeclaration
	// OutputPackage contains information about the package where the generated code will go.
	OutputPackage *PackageDeclaration
	// Module contains info from "go.mod" about the entire module where the service/package is defined.
	Module *ModuleDeclaration
	// Services encapsulates snapshot info for all service interfaces that were defined in the input file.
	Services []*ServiceDeclaration
	// Models encapsulates snapshot info for all service request/response structs that were defined in the input file.
	Models []*ServiceModelDeclaration
	// TypeInfo contains detailed compiler info about the types in your source file.
	TypeInfo *types.Info
}

// ServiceByName looks through "Services" to find the one with the matching interface name.
func (ctx Context) ServiceByName(name string) *ServiceDeclaration {
	for _, service := range ctx.Services {
		if service.Name == name {
			return service
		}
	}
	return nil
}

// ModelByName looks through "Models" to find the one whose method/function name matches 'name'.
func (ctx Context) ModelByName(name string) *ServiceModelDeclaration {
	// These lookups likely happen when we perform lookups for service method parameters. Those are
	// usually pointers (e.g. "*GetPostRequest). The model name we put on the context does not have
	// any sort of pointer identification, so strip that off.
	if strings.HasPrefix(name, "*") {
		name = name[1:]
	}

	for _, model := range ctx.Models {
		if model.Name == name {
			return model
		}
	}
	return nil
}

func (ctx Context) LookupType(typeExpr ast.Expr) (types.Type, error) {
	info, ok := ctx.TypeInfo.Types[typeExpr]
	if !ok {
		return nil, fmt.Errorf("unable to find type info for %v", types.ExprString(typeExpr))
	}
	return info.Type, nil
}

// ServiceDeclaration wrangles all of the information we could grab about the service from the
// interface that defined it.
type ServiceDeclaration struct {
	// Name is the name of the service/interface.
	Name string
	// Version is the (hopefully) semantic version of your API (e.g. 1.2.0). This is NOT the prefix
	// to all routes in the API for the service. It's just an identifier available to code gen tools.
	Version string
	// Gateway contains the configuration HTTP-related options for this service.
	Gateway *GatewayServiceOptions
	// Functions are all of the functions explicitly defined on this service.
	Functions []*ServiceFunctionDeclaration
	// Documentation are all of the comments documenting this service.
	Documentation DocumentationLines
	// Node is the syntax tree object for the interface that described this service.
	Node *ast.Object
}

// InterfaceNode traverses the AST node for the service, returning the properly-cast InterfaceType
// declaration which defined this service.
func (service ServiceDeclaration) InterfaceNode() *ast.InterfaceType {
	return service.Node.
		Decl.(*ast.TypeSpec).
		Type.(*ast.InterfaceType)
}

// FunctionByName fetches the service operation with the given function name. This returns nil when there
// are no functions in this interface/service by that name.
func (service ServiceDeclaration) FunctionByName(name string) *ServiceFunctionDeclaration {
	for _, m := range service.Functions {
		if m.Name == name {
			return m
		}
	}
	return nil
}

// ServiceFunctionDeclaration defines a single operation/function within a service (one of the interface functions).
type ServiceFunctionDeclaration struct {
	// Name is the name of the function defined in the service interface (the function name to call this operation).
	Name string
	// Request contains the details about the model/type/struct for this operation's input/request value.
	Request *ServiceModelDeclaration
	// Response contains the details about the model/type/struct for this operation's output/response value.
	Response *ServiceModelDeclaration
	// Gateway wrangles all of the HTTP-related options for this function (method, path, etc).
	Gateway *GatewayFunctionOptions
	// Documentation are all of the comments documenting this operation.
	Documentation DocumentationLines
	// Node is the syntax tree object that defined this function within the service interface.
	Node *ast.Field
}

// String returns the function signature for this operation for debugging purposes.
func (f ServiceFunctionDeclaration) String() string {
	return fmt.Sprintf("%s(context.Context, %v) (%v, error)",
		f.Name,
		f.Request,
		f.Response,
	)
}

// ServiceModelDeclaration contains information about request/response structs defined in your declaration file.
type ServiceModelDeclaration struct {
	// Name is the name of the type/struct used when defining the request/response value.
	Name string
	// Documentation are all of the comments documenting this operation.
	Documentation DocumentationLines
	// Fields are the individual data attributes on this model/struct.
	Fields FieldDeclarations
	// Type contains the runtime type data about this model.
	Type *FieldType
	// Node is the syntax tree object that defined this type/struct.
	Node *ast.Object
}

// String just returns the model type's name.
func (model ServiceModelDeclaration) String() string {
	return model.Name
}

// FieldDeclarations collects the fields/attributes on a service model.
type FieldDeclarations []*FieldDeclaration

// Empty returns true if there are zero fields in this set.
func (fields FieldDeclarations) Empty() bool {
	return len(fields) == 0
}

// NotEmpty returns true if there is at least one field in this set
func (fields FieldDeclarations) NotEmpty() bool {
	return len(fields) > 0
}

// FieldByName looks up the declaration for the field that matches the given name. This name
// comparison is CASE INSENSITIVE, so "id" will find the field "ID".
func (fields FieldDeclarations) FieldByName(name string) *FieldDeclaration {
	for _, field := range fields {
		if strings.EqualFold(field.Name, name) {
			return field
		}
	}
	return nil
}

// FieldDeclaration describes a single field in a request/response model.
type FieldDeclaration struct {
	// Name the name of the field/attribute.
	Name string
	// Type contains the data type information for this field.
	Type *FieldType
	// Documentation are all of the comments documenting this field.
	Documentation DocumentationLines
	// Node is the syntax tree object where this field was defined.
	Node *ast.Field
}

// FieldType captures a whole bunch of type data related to a single filed on a request/response struct.
type FieldType struct {
	// Name is the fully qualified name/expression for the type (e.g. "uint", "time.Time", "*Foo", "[]byte", etc).
	Name string
	// Pointer indicates if the field's type is a pointer or not.
	Pointer bool
	// Type is the raw, parsed type that is the right-hand-side of the line where the field is defined.
	Type types.Type
	// Underlying peels away all of the type aliases of 'Type' until we get to the raw primitive or struct
	// that truly indicates what this field represents.
	Underlying types.Type
	// Elem is only non-nil for slice/array/chan types. If the slice is this field type, the ElemType
	// describes what the type of each element describes.
	Elem *FieldType
	// Key is only non-nil for tuple/map types where there's some key/value pairing. It describes the type
	// of all of the keys in the collection whereas ElemType describes the value.
	Key *FieldType
	// JSONType is the name of the JS/JSON type that this most naturally maps to (number/string/boolean/object/array).
	JSONType string
}

// ModuleDeclaration contains information about the Go module that the service belongs
// to. This is information scraped from project's "go.mod" file.
type ModuleDeclaration struct {
	// Name is the fully qualified module name (e.g. "github.com/someuser/modulename")
	Name string
	// Directory is the absolute path to the root directory of the module (where go.mod resides).
	Directory string
}

// GoMod returns the absolute path to the "go.mod" file for this module on the system running frodo.
func (module ModuleDeclaration) GoMod() string {
	return filepath.Join(module.Directory, "go.mod")
}

// PackageDeclaration defines the subpackage that the service resides in.
type PackageDeclaration struct {
	// Name is just the raw package name (no path info)
	Name string
	// Import is the fully qualified package name (e.g. "github.com/someuser/modulename/foo/bar/baz")
	Import string
	// Directory is the absolute path to the package.
	Directory string
}

// DocumentationLines represents all of the 'go doc' lines above a type/function/field with all
// of the leading slashes removed.
type DocumentationLines []string

// Trim removes blank doc lines from the front/back of your list of comments.
func (docs DocumentationLines) Trim() DocumentationLines {
	if docs.Empty() {
		return docs
	}
	// We want to be able to trim leading and trailing blank lines in a single pass over the
	// slice, so the first time we encounter a non-empty line that's the first index to keep.
	// The last index will continuously be updated as we go further and find other non-empty
	// lines, so by the time we finish the loop we should have both ends of the valid range.
	first := -1
	last := -1
	for i, line := range docs {
		switch {
		case line == "":
			// don't update anything
		case first < 0:
			// we just found the first non-blank comment
			first = i
			last = i
		default:
			// there's another non-blank comment line further after the first one
			last = i
		}
	}
	return docs[first : last+1]
}

// NotEmpty returns true when there is at least 1 line of documentation/comments.
func (docs DocumentationLines) NotEmpty() bool {
	return len(docs) > 0
}

// Empty returns true when there are no lines of documentation/comments for the service/field/function/etc.
func (docs DocumentationLines) Empty() bool {
	return len(docs) == 0
}

func normalizePathSegment(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	path = strings.TrimSpace(path)
	return path
}

// GatewayServiceOptions contains all of the configurable HTTP-related options for a top-level service.
type GatewayServiceOptions struct {
	// Service is a back-pointer to the service these options correspond to.
	Service *ServiceDeclaration
	// PathPrefix is the optional version/domain prefix for all endpoints in the API (e.g. "v2/").
	PathPrefix string
}

// GatewayFunctionOptions contains all of the configurable HTTP-related options for a single
// function within your service (e.g. method, path, etc).
type GatewayFunctionOptions struct {
	// Function is a back-pointer to the service function these options correspond to.
	Function *ServiceFunctionDeclaration
	// Method indicates if the RPC gateway should use a GET, POST, etc when exposing this operation via HTTP.
	Method string
	// Path defines the URL pattern to provide to the gateway's router/mux to access this operation.
	Path string
	// Status indicates what success status code the gateway should use when responding via HTTP (e.g. 200, 202, etc)
	Status int
}

// SupportsBody returns true when the method is either POST, PUT, or PATCH; the HTTP methods
// where we expect you to feed request data via the request body rather than query string.
func (opts GatewayFunctionOptions) SupportsBody() bool {
	method := strings.ToUpper(opts.Method)
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

// PathParameters looks at all of the ":xxx" path parameters in HTTPPath and returns the fields on
// the request struct that will be bound by those values at runtime. For instance, if the path
// was "/user/:userID/address/:addressID", this will return a 2-element slice containing the request's
// UserID and AddressID fields.
func (opts GatewayFunctionOptions) PathParameters() GatewayParameters {
	var results GatewayParameters
	for _, segment := range strings.Split(opts.Path, "/") {
		if !strings.HasPrefix(segment, ":") {
			continue
		}

		paramName := segment[1:]
		field := opts.Function.Request.Fields.FieldByName(paramName)
		if field == nil {
			continue
		}

		results = append(results, &GatewayParameter{
			Name:  paramName,
			Field: field,
		})
	}
	return results
}

func (opts GatewayFunctionOptions) QueryParameters() GatewayParameters {
	var results GatewayParameters

	// If you're doing a POST/PUT/PATCH, we expect every value to come from either
	// the body or the path, not the query string.
	if opts.SupportsBody() {
		return results
	}

	pathParams := opts.PathParameters()

	for _, field := range opts.Function.Request.Fields {
		// Exclude any fields that will be bound using path parameters.
		if pathParams.ByName(field.Name) != nil {
			continue
		}

		results = append(results, &GatewayParameter{
			Name:  field.Name,
			Field: field,
		})
	}
	return results
}

// GatewayParameters is an overlay of a service function's path and request type/field info. It helps you
// indicate how a given field will be bound when handling incoming requests (e.g. path params vs query params).
type GatewayParameters []*GatewayParameter

// ByName locates the parameter with the given name. This is a case-insensitive search.
func (params GatewayParameters) ByName(name string) *GatewayParameter {
	for _, param := range params {
		if strings.EqualFold(param.Name, name) {
			return param
		}
	}
	return nil
}

// Empty returns true when there are zero parameters defined in this set.
func (params GatewayParameters) Empty() bool {
	return len(params) == 0
}

// NotEmpty returns true when there is at least one parameter mapping defined.
func (params GatewayParameters) NotEmpty() bool {
	return !params.Empty()
}

// GatewayParameter defines how a path/query parameter will be bound to a field in your request struct.
type GatewayParameter struct {
	// Name is the identifier of the path param (e.g "id" in "/user/:id") or query string value that
	// will be bound to the Field.
	Name string
	// Field indicates which model attribute will be populated when this parameter goes
	// through the request binder.
	Field *FieldDeclaration
}

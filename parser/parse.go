package parser

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/monadicstack/frodo/internal/implements"
	"github.com/monadicstack/frodo/internal/naming"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

// DefaultServiceVersion defines the version we'll assign to all parsed services if they do
// not have the VERSION doc option.
const DefaultServiceVersion = "0.0.1"

// ErrNoServices is the error returned when your input file does not contain any "XyzService" interfaces.
var ErrNoServices = fmt.Errorf("file does not contain any service interfaces")

// ErrMultipleServices is the error returned when you define multiple "XyzService" interfaces in one file.
var ErrMultipleServices = fmt.Errorf("do not define multiple services in a single file")

// ErrMultiplePackages is the error returned when you try to parse multiple packages for types at once.
var ErrMultiplePackages = fmt.Errorf("multiple packages defined in input path; should be one")

// ErrMissingGoMod is the error returned when the project we're parsing does not have a 'go.mod' file in it.
var ErrMissingGoMod = fmt.Errorf("unable to find 'go.mod' for project")

// ErrTypeNotStructPointer is the error returned when the request/response value is not a struct pointer.
var ErrTypeNotStructPointer = fmt.Errorf("not a pointer to a struct type")

// ErrTypeNotError is the error returned when the second return value of an operation is not an error.
var ErrTypeNotError = fmt.Errorf("not the type 'error'")

// ErrTypeNotContext is the error returned when the first param of an operation is not a context.Context.
var ErrTypeNotContext = fmt.Errorf("not the type 'context.Context'")

// ErrTypeNotTwoParams is the error for when your function signature doesn't accept two parameters.
var ErrTypeNotTwoParams = fmt.Errorf("must have two params")

// ErrTypeNotTwoReturns is the error for when your function signature doesn't return two values.
var ErrTypeNotTwoReturns = fmt.Errorf("must have two return values")

// ParseFile parses a source code file containing a service interface declaration as well as the
// structs for the request/response inputs and outputs. It will aggregate all of the services/ops/models
// described in the source code in a much more simple/direct Context.
//
// The resulting Context contains all of the information from the source code that we need to generate
// clients/gateways for the service(s). It will also be used as the input value when evaluating any
// of our artifact templates.
func ParseFile(inputPath string) (*Context, error) {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, inputPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("unable to parse go file: %s: %w", inputPath, err)
	}

	absolutePath, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse go file: %s: %w", inputPath, err)
	}

	ctx := &Context{
		FileSet:      fileSet,
		File:         file,
		Path:         inputPath,
		AbsolutePath: absolutePath,
	}

	if ctx.Module, err = ParseModuleInfo(ctx); err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", inputPath, err)
	}
	if ctx.InputPackage, ctx.OutputPackage, err = ParsePackageInfo(ctx); err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", inputPath, err)
	}
	if ctx.Documentation, ctx.Tags, err = ParseDocumentation(ctx); err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", inputPath, err)
	}
	if ctx.RawTypes, ctx.Types, err = ParseTypes(ctx); err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", inputPath, err)
	}
	if ctx.Service, err = ParseService(ctx); err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", inputPath, err)
	}
	return ctx, nil
}

// ParseTypes runs the syntax tree through the "go/types" processor so that we get detailed
// type information on all of the structs/types we defined, their fields, and the parameters/outputs
// of our service functions.
func ParseTypes(ctx *Context) (*packages.Package, TypeRegistry, error) {
	config := &packages.Config{
		Tests: false,
		Mode:  packages.NeedDeps | packages.NeedName | packages.NeedSyntax | packages.NeedTypes,
	}

	loadedPackages, err := packages.Load(config, ctx.Path)
	if err != nil {
		return nil, nil, err
	}
	if len(loadedPackages) != 1 {
		return nil, nil, ErrMultiplePackages
	}

	targetPackage := loadedPackages[0]
	targetScope := targetPackage.Types.Scope()
	registry := NewTypeRegistry()

	// Iterate our top level type definitions and then recursively iterate their fields' types to populate
	// our entire type registry of every single type that this service requires.
	for _, scopeKey := range targetScope.Names() {
		t := targetScope.Lookup(scopeKey).Type()
		typeDeclaration := registerType(ctx, registry, t)
		ApplyTypeDocumentation(ctx, typeDeclaration)
	}
	return targetPackage, registry.WithoutInvalid(), nil
}

func registerType(ctx *Context, registry TypeRegistry, t types.Type) *TypeDeclaration {
	// We've already added this type to the registry, so avoid circular recursion.
	if entry, ok := registry.Lookup(t); ok {
		return entry
	}

	// Add it to the registry before iterating any struct fields so that if one of its fields is this type, we
	// don't infinitely try to register it over and over (the if check above). A case for this might be like a linked
	// list where a Node struct might have a pointer to the next Node.
	name := t.String()
	name = naming.NoImport(name)
	name = naming.NoPointer(name)
	name = naming.CleanPrefix(name)
	typeDeclaration := registry.Register(&TypeDeclaration{Name: name, Type: t})

	registerTypeEntry(ctx, registry, typeDeclaration, t)
	return ApplyTypeDocumentation(ctx, typeDeclaration)
}

func registerTypeEntry(ctx *Context, registry TypeRegistry, entry *TypeDeclaration, t types.Type) {
	switch tt := t.(type) {
	case *types.Pointer:
		// We track "pointer-ness" on fields whose type is a pointer, not on the type itself, so just apply
		// the pointer type's information to the core type entry.
		registerTypeEntry(ctx, registry, entry, tt.Elem())

	case *types.Struct:
		// Recursively parse the type information for all the field members of the struct.
		entry.Kind = reflect.Struct
		parseStructFields(ctx, registry, entry, tt)

	case *types.Named:
		// The "Named" type doesn't actually have any meaningful information. For example, if the declaration
		// is "type Foo []Bar", the Named type is "Foo", but we need to fill Foo's entry w/ information stored
		// on the underlying type, "[]Bar".
		registerTypeEntry(ctx, registry, entry, tt.Underlying())

		// Check to see if any/all of our raw file interfaces are implemented. This will serve as helper data for the
		// client/gateway generators to know when a response should be treated as JSON (default) or raw bytes.
		entry.Implements.ContentReader = implements.Method(tt, "Content", nil, []string{"io.ReadCloser"})
		entry.Implements.ContentTypeReader = implements.Method(tt, "ContentType", nil, []string{"string"})
		entry.Implements.ContentFileNameReader = implements.Method(tt, "ContentFileName", nil, []string{"string"})
		entry.Implements.ContentWriter = implements.Method(tt, "SetContent", []string{"io.ReadCloser"}, nil)
		entry.Implements.ContentTypeWriter = implements.Method(tt, "SetContentType", []string{"string"}, nil)
		entry.Implements.ContentFileNameWriter = implements.Method(tt, "SetContentFileName", []string{"string"}, nil)

	case *types.Array:
		entry.Basic = entry.Type == t
		entry.Kind = reflect.Array
		entry.Elem = registerType(ctx, registry, tt.Elem())

	case *types.Slice:
		entry.Basic = entry.Type == t
		entry.Kind = reflect.Slice
		entry.Elem = registerType(ctx, registry, tt.Elem())

	case *types.Map:
		entry.Basic = entry.Type == t
		entry.Kind = reflect.Map
		entry.Key = registerType(ctx, registry, tt.Key())
		entry.Elem = registerType(ctx, registry, tt.Elem())

	case *types.Basic:
		// Our default registry should already have all of the basic types pre-populated, so just use that.
		entry.Kind = registry[t.String()].Kind

	case *types.Interface:
		entry.Kind = reflect.Interface

	default:
		// We don't allow channels or function types to be considered "valid" types on your request/response
		// struct fields, so we're going to weed these out.
		entry.Kind = reflect.Invalid
		return
	}
}

func parseStructFields(ctx *Context, registry TypeRegistry, model *TypeDeclaration, structType *types.Struct) {
	for _, structField := range flattenedStructFields(structType) {
		fieldDecl := parseStructField(ctx, registry, model, structField)
		if fieldDecl == nil {
			continue
		}
		model.Fields = append(model.Fields, fieldDecl)
	}
}

func parseStructField(ctx *Context, registry TypeRegistry, model *TypeDeclaration, structField *types.Var) *FieldDeclaration {
	fieldType := registerType(ctx, registry, structField.Type())
	if fieldType.Kind == reflect.Invalid {
		return nil
	}

	fieldDecl := &FieldDeclaration{
		Name:       structField.Name(),
		ParentType: model,
		Type:       fieldType,
		Pointer:    pointerType(structField.Type()),
	}
	fieldDecl.Binding = ParseBindingOptions(ctx, fieldDecl, structField)
	return ApplyFieldDocumentation(ctx, fieldDecl)
}

// ParsePackageInfo overlays your project's "go.mod" file and your input file/path to figure
// out the fully qualified package info for the input service. We'll then apply our conventions
// to construct info about the output package where we'll put all of our output artifacts.
func ParsePackageInfo(ctx *Context) (input *PackageDeclaration, output *PackageDeclaration, err error) {
	moduleDir, _ := filepath.Abs(ctx.Module.Directory)
	packageDir, _ := filepath.Abs(filepath.Dir(ctx.Path))
	packageDirRelative := strings.TrimPrefix(packageDir, moduleDir)
	packageName := ctx.File.Name.Name

	input = &PackageDeclaration{
		Name:      packageName,
		Import:    filepath.Join(ctx.Module.Name, packageDirRelative),
		Directory: filepath.Dir(ctx.Path),
	}
	output = &PackageDeclaration{
		Name:      packageName,
		Import:    filepath.Join(input.Import, "gen"),
		Directory: filepath.Join(input.Directory, "gen"),
	}
	return input, output, nil
}

// ParseModuleInfo cherry picks a tiny bit of info from your "go.mod" file that we use
// in processing your services.
func ParseModuleInfo(ctx *Context) (*ModuleDeclaration, error) {
	inputFilePath, err := filepath.Abs(ctx.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to determine absolute path: %w", err)
	}

	// Look in the input file's directory (an all of its parents/ancestors) for the "go.mod" file.
	goModPath, err := FindGoDotMod(filepath.Dir(inputFilePath))
	if err != nil {
		return nil, err
	}
	// Read/parse the "go.mod" file so we can extract the module/package info we need.
	goModData, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	goModFile, err := modfile.Parse(goModPath, goModData, nil)
	if err != nil {
		return nil, err
	}

	return &ModuleDeclaration{
		Name:      goModFile.Module.Mod.Path,
		Directory: filepath.Dir(goModPath),
	}, nil
}

// FindGoDotMod starts in the current directory provided and recursively checks
// parent directories until it encounters a "go.mod" file. When it does, this will
// return a path to the file. You'll receive an error if we can't find a "go.mod"
// file or the input is not a valid directory.
func FindGoDotMod(dirName string) (string, error) {
	if dirName == "" || dirName == "/" {
		return "", ErrMissingGoMod
	}

	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return "", fmt.Errorf("unable to find 'go.mod': %w", err)
	}

	for _, file := range files {
		if file.Name() == "go.mod" {
			return filepath.Join(dirName, file.Name()), nil
		}
	}
	return FindGoDotMod(filepath.Dir(dirName))
}

// ParseDocumentation runs go/doc parsing on your input file to extract all of your documentation, comments, and
// struct field tags. It returns 2 specialized lookup maps; one for doc comments and one for the struct field tags.
// The keys to these maps are based on the names of the thing whose docs/tags you want; either "SERVICE",
// "SERVICE.FUNCTION", "MODEL", or "MODEL.FIELD".
func ParseDocumentation(ctx *Context) (Documentation, Tags, error) {
	packageDocs, err := doc.NewFromFiles(ctx.FileSet, []*ast.File{ctx.File}, ctx.Module.Name)
	if err != nil {
		return nil, nil, err
	}

	/*
	 * What's going on here? We're collecting all of the GoDoc comments on the services, structs, functions,
	 * and fields defined in your input file and storing them in an easier lookup mechanism for our purposes. The
	 * packageDocs value is a tree we can traverse to find any doc we want, but it's inefficient since we look up
	 * the corresponding comments for a service/function/etc on demand. Instead, we create an O(1) lookup map that
	 * uses the names of the types/functions/etc to find the appropriate comments. For instance, when all is said and
	 * done, we want a map that looks like this:
	 *
	 *     "UserService":          "UserService provides all operations on...",
	 *     "UserService.GetByID":  "GetByID finds a user given their unique id",
	 *     "UserService.Search":   "Search finds all users matching the specified criteria.",
	 *     "SearchRequest" :       "SearchRequest contains all of the filtering options for...",
	 *     "SearchRequest.Text":   "Text limits the search to only include users with these tokens...",
	 *     "SearchRequest.Limit":  "Limit is the maximum number of users the search will return.",
	 *     "SearchRequest.Offset": "Offset handles paging by skipping...",
	 *     ...
	 *
	 * The keys are either "SERVICE", "SERVICE.FUNCTION", "MODEL", or "MODEL.FIELD". We traverse the tree and build
	 * this flat structure so that we can do easier lookups later.
	 *
	 * As for the 'Tags', it's the exact same process; turn the tree into a flat map so we can find the `json` tags
	 * on fields later on. True, tags arent' *technically* documentation, but it is where we can easily grab this
	 * information. The type tree from "packages.Load()" does not have this data, so while we're traversing the AST
	 * for documentation, we can grab the tag info, too.
	 */
	docs := Documentation{}
	tags := Tags{}

	// First iterate through the service interfaces and capture the interface/function docs.
	for _, service := range packageDocs.Types {
		interfaceNode, ok := toInterfaceTypeNode(service)
		if !ok {
			continue
		}

		docs.Set(service.Name, service.Doc)
		for _, function := range interfaceNode.Methods.List {
			docs.Set(service.Name, fieldName(function), toCommentText(function.Doc))
		}
	}

	// Now iterate the models to capture the model/field docs.
	for _, model := range packageDocs.Types {
		structNode, ok := toStructTypeNode(model)
		if !ok {
			continue
		}

		docs.Set(model.Name, model.Doc)
		for _, field := range structNode.Fields.List {
			docs.Set(model.Name, fieldName(field), toCommentText(field.Doc))
			tags.Set(model.Name, fieldName(field), field.Tag)
		}
	}

	return docs, tags, nil
}

// toCommentText is a nil-safe way to extract the raw GoDoc comment string from a node's comment group. Should
// you provide a nil group (i.e. the node doesn't have comments), this will just return "".
func toCommentText(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return group.Text()
}

// toStructTypeNode accepts a documentation tree's type node and returns the underlying AST struct node for
// it. If documentation node is not part of a struct, this will return nil/false.
func toStructTypeNode(t *doc.Type) (*ast.StructType, bool) {
	typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil, false
	}
	structNode, ok := typeSpec.Type.(*ast.StructType)
	return structNode, ok
}

// toStructTypeNode accepts a documentation tree's type node and returns the underlying AST interface node for
// it. If documentation node is not part of an interface, this will return nil/false.
func toInterfaceTypeNode(t *doc.Type) (*ast.InterfaceType, bool) {
	typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil, false
	}
	interfaceNode, ok := typeSpec.Type.(*ast.InterfaceType)
	return interfaceNode, ok
}

// findServiceInterface fetches only the Interface type nodes from the scope/AST that look like service declarations.
// It will return the name of the interface and the AST node type info for the service interface it finds. You'll
// receive a non-nil error if there are no service interfaces or more than 1 service interface.
func findServiceInterface(ctx *Context) (string, *types.Interface, error) {
	var serviceName string
	var serviceInterface *types.Interface

	for _, name := range ctx.Scope().Names() {
		nextService, ok := toServiceInterface(ctx.Scope().Lookup(name))
		if !ok {
			continue
		}
		if serviceInterface != nil {
			return "", nil, ErrMultipleServices
		}
		serviceInterface = nextService
		serviceName = name
	}

	if serviceInterface == nil {
		return "", nil, ErrNoServices
	}
	return serviceName, serviceInterface, nil
}

// toServiceInterface accepts a 'type' from the packages type tree and returns the raw interface data for it
// if and only if it meets our criteria for being a "service interface"
//
// * It follows the naming convention "FooBarService" (i.e. ends with "Service")
// * It is an exported type
// * The 'type' is interface.
//
// Any type that doesn't meet all of these criteria will receive "nil, false" back.
func toServiceInterface(typeObj types.Object) (*types.Interface, bool) {
	// Enforce the naming convention that services end w/ the word "Service"
	if !strings.HasSuffix(typeObj.Name(), "Service") {
		return nil, false
	}
	if !typeObj.Exported() {
		return nil, false
	}
	return underlyingInterface(typeObj.Type())
}

// toModelStruct accepts a 'type' from the packages type tree and returns the raw struct data for it
// if and only if it meets our criteria for being a valid request/response model:
//
// * It is an exported type
// * The 'type' is struct.
//
// Pretty simple. Any type that doesn't meet all those criteria will receive "nil, false" back.
func toModelStruct(typeObj types.Object) (*types.Struct, bool) {
	if !typeObj.Exported() {
		return nil, false
	}
	return underlyingStruct(typeObj.Type())
}

// ParseService looks for 'type XxxService interface' declarations and extracts all
// service/operation info from it that we need to generate our artifacts. This operation will
// fail if you have multiple service interfaces in this file.
func ParseService(ctx *Context) (*ServiceDeclaration, error) {
	var err error

	// First, make sure we have one and only one service defined in this file.
	serviceName, serviceInterface, err := findServiceInterface(ctx)
	if err != nil {
		return nil, err
	}

	// Now scrape that one and only service's data into a declaration instance.
	service := ApplyServiceDocumentation(ctx, &ServiceDeclaration{
		Name:    serviceName,
		Version: DefaultServiceVersion,
		Gateway: &GatewayServiceOptions{},
	})
	service.Gateway.Service = service

	service.Functions, err = ParseServiceFunctions(ctx, service, serviceInterface)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// ParseServiceFunctions creates function declarations for all methods on the service interface.
func ParseServiceFunctions(ctx *Context, service *ServiceDeclaration, interfaceType *types.Interface) ([]*ServiceFunctionDeclaration, error) {
	var functions []*ServiceFunctionDeclaration
	for i := 0; i < interfaceType.NumMethods(); i++ {
		function, err := ParseServiceFunction(ctx, service, interfaceType.Method(i))
		if err != nil {
			return nil, err
		}

		functions = append(functions, function)
	}
	return functions, nil
}

// ParseServiceFunction captures the information for a single function on a service. This includes all of the
// doc options that configure the gateway stuff.
func ParseServiceFunction(ctx *Context, service *ServiceDeclaration, funcType *types.Func) (*ServiceFunctionDeclaration, error) {
	function := &ServiceFunctionDeclaration{
		Name:    funcType.Name(),
		Service: service,
		Gateway: &GatewayFunctionOptions{
			Status: http.StatusOK,
			Method: http.MethodPost,
			Path:   "/" + service.Name + "." + funcType.Name(),
		},
	}
	function.Gateway.Function = function

	signature, ok := funcType.Type().(*types.Signature)
	if !ok {
		return nil, fmt.Errorf("%s.%s(): not a function signature type", service.Name, function.Name)
	}

	// Check to make sure that we have 2 parameters and 2 return values.
	if signature.Params().Len() != 2 {
		return nil, fmt.Errorf("%s.%s(): %w", service.Name, function.Name, ErrTypeNotTwoParams)
	}
	if signature.Results().Len() != 2 {
		return nil, fmt.Errorf("%s.%s(): %w", service.Name, function.Name, ErrTypeNotTwoReturns)
	}

	param1 := signature.Params().At(0)
	param2 := signature.Params().At(1)
	result1 := signature.Results().At(0)
	result2 := signature.Results().At(1)

	// Make sure that the two inputs are a context.Context and a request struct.
	if !validMethodParam1(ctx, param1) {
		return nil, fmt.Errorf("%s.%s(): param 1: %w", service.Name, function.Name, ErrTypeNotContext)
	}
	if !validMethodParam2(ctx, param2) {
		return nil, fmt.Errorf("%s.%s(): param 2: %w", service.Name, function.Name, ErrTypeNotStructPointer)
	}

	// Make sure that the two return values are a response struct and an error
	if !validMethodReturnValue1(ctx, result1) {
		return nil, fmt.Errorf("%s.%s(): return value 1: %w", service.Name, function.Name, ErrTypeNotStructPointer)
	}
	if !validMethodReturnValue2(ctx, result2) {
		return nil, fmt.Errorf("%s.%s(): return value 2: %w", service.Name, function.Name, ErrTypeNotError)
	}

	// We're enforcing a convention that you define your request/response structs in the same file as the
	// services that they correspond to. Even if you want to share common types across services, that's fine,
	// but you need to define an alias or a new type where the common type is embedded in that file.
	if function.Request, _ = ctx.Types.Lookup(param2.Type()); function.Request == nil {
		return nil, fmt.Errorf("%s(): request struct must be defined in %s", function.Name, ctx.Path)
	}
	if function.Response, _ = ctx.Types.Lookup(result1.Type()); function.Response == nil {
		return nil, fmt.Errorf("%s(): response struct must be defined in %s", function.Name, ctx.Path)
	}

	ApplyFunctionDocumentation(ctx, function)
	return function, nil
}

func flattenedStructFields(structType *types.Struct) []*types.Var {
	var fields []*types.Var
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		if !field.Exported() {
			continue
		}
		if !field.Embedded() {
			fields = append(fields, field)
			continue
		}

		// Embedded struct support falls into one of these two buckets for us.
		//
		//     type Request {
		//         Record
		//         Name
		//     }
		//     type Record struct {
		//         ID string
		//     }
		//     type Name string
		//
		// The embedded 'Record' field is another struct, so we need to recursively grab its fields
		// and include them in this flattened list. The embedded 'Name' field, however, is just a standalone
		// field, so include it like a normal field.
		embeddedStruct, ok := underlyingStruct(field.Type())
		if !ok {
			fields = append(fields, field)
			continue
		}
		fields = append(fields, flattenedStructFields(embeddedStruct)...)
	}
	return fields
}

// ParseBindingOptions looks at the `json` tags of the given struct field and returns this field's binding
// configuration. It indicates whether the field should be left out of JSON marshaling and what field name to
// use when going to/from JSON format. If the field has no `json` tag, you will get a set of binding options
// representing the default values (i.e. include the field and use its exact name).
func ParseBindingOptions(ctx *Context, field *FieldDeclaration, fieldVar *types.Var) *FieldBindingOptions {
	options := &FieldBindingOptions{
		Omit: false,
		Name: varName(fieldVar),
	}

	// The field doesn't have a 'json' tag assigned or they weirdly defined `json:""`, then
	// the default binding options reign supreme.
	tag := ctx.Tags.ForField(field).Get("json")
	if tag == "" {
		return options
	}

	// We don't care about 'omitempty' or anything other than the remapped name. The
	// runtime binder cares, but not the syntax parser.
	switch name := strings.Split(tag, ",")[0]; name {
	case "-":
		options.Omit = true
		return options
	default:
		options.Name = name
		return options
	}
}

// The first param to all service functions should be a standard "context.Context"
func validMethodParam1(_ *Context, param *types.Var) bool {
	// Look up the real type from the Go parser rather than reading the type info directly
	// off of the 'param'. This ensures that if you aliased the "context" package, we can
	// still properly identify it. If you aliased it to "foo" your function might look like this:
	//
	//     GetByID(foo.Context, *GetByIDRequest) (*GetByIDResponse, error)
	//
	// According to the 'param' the type name is "foo.Context". We have no idea what "foo" is, so
	// we go back to the Go parser's type info table to identify the un-aliased type name
	// which is "context.Context"; and that's what we look for.
	return param.Type().String() == "context.Context"
}

// The second parameter should be a pointer to your "request struct".
func validMethodParam2(_ *Context, param *types.Var) bool {
	if _, ok := underlyingPointer(param.Type()); !ok {
		return false
	}
	if _, ok := underlyingStruct(param.Type()); !ok {
		return false
	}
	return true
}

// The first return value should be a pointer to your "response struct".
func validMethodReturnValue1(ctx *Context, param *types.Var) bool {
	// It has the same semantics - must be a pointer to a struct.
	return validMethodParam2(ctx, param)
}

// The second return value should always just be an error.
func validMethodReturnValue2(_ *Context, param *types.Var) bool {
	return param.Type().String() == "error"
}

func underlyingInterface(t types.Type) (*types.Interface, bool) {
	switch typed := t.(type) {
	case *types.Interface:
		return typed, true
	case *types.Named:
		return underlyingInterface(t.Underlying())
	case *types.Pointer:
		return underlyingInterface(typed.Elem())
	default:
		return nil, false
	}
}

func underlyingPointer(t types.Type) (*types.Pointer, bool) {
	switch typed := t.(type) {
	case *types.Pointer:
		return typed, true
	case *types.Named:
		return underlyingPointer(t.Underlying())
	default:
		return nil, false
	}
}

func underlyingStruct(t types.Type) (*types.Struct, bool) {
	switch typed := t.(type) {
	case *types.Struct:
		return typed, true
	case *types.Named:
		return underlyingStruct(t.Underlying())
	case *types.Pointer:
		return underlyingStruct(typed.Elem())
	default:
		return nil, false
	}
}

// parseHTTPStatus is just a strconv.ParseInt that parses the right hand side of an "HTTP 202"
// looking comment. If we can't parse it as a number for any reason, we'll default to 200.
func parseHTTPStatus(statusText string) int {
	statusText = strings.TrimSpace(statusText)
	status, err := strconv.ParseInt(statusText, 10, 64)
	if err != nil {
		return http.StatusOK
	}
	return int(status)
}

// ApplyServiceDocumentation takes the documentation comment block above your interface type
// declaration and applies them to the service snapshot, parsing all Doc Options in the process.
func ApplyServiceDocumentation(ctx *Context, service *ServiceDeclaration) *ServiceDeclaration {
	if ctx == nil || service == nil {
		return service
	}

	for _, line := range ctx.Documentation.ForService(service) {
		switch {
		case strings.HasPrefix(line, "PATH "):
			service.Gateway.PathPrefix = normalizePath(line[5:])
		case strings.HasPrefix(line, "PREFIX "):
			service.Gateway.PathPrefix = normalizePath(line[7:])
		case strings.HasPrefix(line, "VERSION "):
			service.Version = strings.TrimSpace(line[8:])
		default:
			service.Documentation = append(service.Documentation, line)
		}
	}
	service.Documentation = service.Documentation.Trim()
	return service
}

// ApplyFunctionDocumentation takes the documentation comment block above your interface function
// declaration and applies them to the function snapshot, parsing all Doc Options in the process.
func ApplyFunctionDocumentation(ctx *Context, function *ServiceFunctionDeclaration) {
	if ctx == nil || function == nil {
		return
	}

	// Notice that "OPTIONS /" is not one of the cases. That's by design. When the gateway
	// registers your POST operation (or whatever method), we're actually going to register
	// that method AND an OPTIONS route for you. By default, the OPTIONS route will simply
	// reject the request (i.e. no default CORS). If bring your own CORS middleware to the
	// party it will respond affirmatively before the rejection. There's more info in the
	// comments of gateway.New() that describes why we need this limitation for now.
	for _, line := range ctx.Documentation.ForFunction(function) {
		switch {
		case strings.HasPrefix(line, "GET "):
			function.Gateway.Method = http.MethodGet
			function.Gateway.Path = normalizePath(line[4:])
		case strings.HasPrefix(line, "PUT "):
			function.Gateway.Method = http.MethodPut
			function.Gateway.Path = normalizePath(line[4:])
		case strings.HasPrefix(line, "POST "):
			function.Gateway.Method = http.MethodPost
			function.Gateway.Path = normalizePath(line[5:])
		case strings.HasPrefix(line, "PATCH "):
			function.Gateway.Method = http.MethodPatch
			function.Gateway.Path = normalizePath(line[6:])
		case strings.HasPrefix(line, "DELETE "):
			function.Gateway.Method = http.MethodDelete
			function.Gateway.Path = normalizePath(line[7:])
		case strings.HasPrefix(line, "HEAD "):
			function.Gateway.Method = http.MethodHead
			function.Gateway.Path = normalizePath(line[5:])
		case strings.HasPrefix(line, "HTTP "):
			function.Gateway.Status = parseHTTPStatus(line[5:])
		default:
			function.Documentation = append(function.Documentation, line)
		}
	}
	function.Documentation = function.Documentation.Trim()
}

// ApplyTypeDocumentation takes the documentation comment block above your struct/alias type
// declaration and applies them to the model snapshot, parsing all Doc Options in the process.
func ApplyTypeDocumentation(ctx *Context, t *TypeDeclaration) *TypeDeclaration {
	if t == nil {
		return t
	}
	t.Documentation = ctx.Documentation.ForType(t).Trim()
	return t
}

// ApplyFieldDocumentation takes the documentation comment block above your struct field
// declaration and applies them to the model snapshot, parsing all Doc Options in the process.
func ApplyFieldDocumentation(ctx *Context, field *FieldDeclaration) *FieldDeclaration {
	if field == nil {
		return field
	}
	field.Documentation = ctx.Documentation.ForField(field).Trim()
	return field
}

// fieldName returns the actual field name that should be used for this attribute within a struct.
func fieldName(field *ast.Field) string {
	// This is an embedded field, so the name is the raw name of the type.
	if len(field.Names) == 0 {
		return naming.NoPointer(naming.NoPackage(types.ExprString(field.Type)))
	}
	return field.Names[0].Name
}

// pointerType determines if 't' represents a pointer type; either directly or it's an alias of one... or
// an alias of an alias of one, etc.
func pointerType(t types.Type) bool {
	if _, ok := t.(*types.Pointer); ok {
		return true
	}

	// You've hit some "root" type like a basic type or something like that, so it's definitely not a pointer.
	underlying := t.Underlying()
	if t == underlying {
		return false
	}
	return pointerType(underlying)
}

// varName takes a struct attribute and returns the simple name that we'll use in our context. This
// handles standard, named fields as well as embedded fields.
func varName(v *types.Var) string {
	if v.Embedded() {
		name := v.Type().String()
		name = naming.NoPackage(name)
		name = naming.NoPointer(name)
		return name
	}
	return v.Name()
}

// normalizePath strips off leading/trailing whitespace, trailing slashes, and ensures that
// your path absolutely begins with a leading slash.
func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	path = strings.TrimSpace(path)
	return "/" + path
}

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
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

// DefaultServiceVersion defines the version we'll assign to all parsed services if they do
// not have the VERSION doc option.
const DefaultServiceVersion = "0.0.1"

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
		return nil, fmt.Errorf("[%s] unable to parse go file: %w", inputPath, err)
	}

	absolutePath, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, fmt.Errorf("[%s] unable to parse go file: %w", inputPath, err)
	}

	ctx := &Context{
		FileSet:      fileSet,
		File:         file,
		Path:         inputPath,
		AbsolutePath: absolutePath,
	}

	if ctx.Module, err = ParseModuleInfo(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.InputPackage, ctx.OutputPackage, err = ParsePackageInfo(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.TypeInfo, err = ParseTypeInformation(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.Documentation, ctx.Tags, err = ParseDocumentation(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.Models, err = ParseModels(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.Services, err = ParseServices(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}

	if len(ctx.Services) == 0 {
		return nil, fmt.Errorf("[%s]: input does not contain any service interfaces", inputPath)
	}
	return ctx, nil
}

// ParseTypeInformation runs the syntax tree through the "go/types" processor so that we get detailed
// type information on all of the structs/types we defined, their fields, and the parameters/outputs
// of our service functions.
func ParseTypeInformation(ctx *Context) (*packages.Package, error) {
	config := &packages.Config{
		Tests: false,
		Mode:  packages.NeedDeps | packages.NeedName | packages.NeedSyntax | packages.NeedTypes,
	}

	loadedPackages, err := packages.Load(config, ctx.Path)
	if err != nil {
		return nil, err
	}
	if len(loadedPackages) != 1 {
		return nil, fmt.Errorf("multiple packages defined in input path; should be one")
	}
	return loadedPackages[0], nil
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
		return "", fmt.Errorf("unable to find 'go.mod' for project")
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

// ParseFieldType looks at the Go parser's type information for a given model attribute and extracts
// all of the various info we need to get a complete picture of the type and how to unravel any
// aliasing that might be going on.
func ParseFieldType(ctx *Context, t types.Type) *FieldType {
	fieldType := &FieldType{
		Name:       typeName(t),
		Type:       t,
		Underlying: underlyingType(t),
	}

	switch fieldTypeType := fieldType.Type.(type) {
	case *types.Pointer:
		fieldType.Pointer = true
		fieldType.Type = fieldTypeType.Elem()
		fieldType.Underlying = underlyingType(fieldType.Type)
	case *types.Array:
		fieldType.Elem = ParseFieldType(ctx, fieldTypeType.Elem())
	case *types.Slice:
		fieldType.Elem = ParseFieldType(ctx, fieldTypeType.Elem())
	case *types.Chan:
		fieldType.Elem = ParseFieldType(ctx, fieldTypeType.Elem())
	case *types.Map:
		fieldType.Key = ParseFieldType(ctx, fieldTypeType.Key())
		fieldType.Elem = ParseFieldType(ctx, fieldTypeType.Elem())
	}

	fieldType.JSONType = toJSON(ctx, fieldType.Underlying)
	return fieldType
}

// toJSON maps the raw Go type to the closest JSON equivalent type (e.g. uint32 -> "number").
func toJSON(ctx *Context, t types.Type) string {
	switch raw := t.(type) {
	case *types.Pointer:
		return toJSON(ctx, raw.Elem())
	case *types.Array, *types.Slice:
		return "array"
	}

	jsonType, ok := jsonTypeMapping[t.String()]
	if ok {
		return jsonType
	}
	return "object"
}

var jsonTypeMapping = map[string]string{
	"string":    "string",
	"bool":      "boolean",
	"rune":      "number",
	"byte":      "number",
	"int":       "number",
	"int8":      "number",
	"int16":     "number",
	"int32":     "number",
	"int64":     "number",
	"uint":      "number",
	"uint8":     "number",
	"uint16":    "number",
	"uint32":    "number",
	"uint64":    "number",
	"uintptr":   "number",
	"float32":   "number",
	"float64":   "number",
	"time.Time": "string",

	// Special case for 'time.Time' when you've dug down to the underlying root type.
	"struct{wall uint64; ext int64; loc *time.Location}": "string",
}

func underlyingType(fieldType types.Type) types.Type {
	name := fieldType.String()

	// In an idea world we'd know if the type implemented MarshalJSON/UnmarshalJSON so
	// that we know if the RPC transport for struct types is an object or some other type.
	// In most cases, the transport for "time.Time" is an ISO8601 string, so we need to short
	// circuit the recursion and stop here rather than going deeper to the "struct{}" type
	// which is not how time is marshaled in Go.
	if name == "time.Time" || name == "*time.Time" {
		return fieldType
	}

	pointer, ok := fieldType.(*types.Pointer)
	if ok {
		return pointer.Elem()
	}

	underlying := fieldType.Underlying()
	if underlying != fieldType {
		return underlyingType(underlying)
	}
	return fieldType
}

// ParseDocumentation runs go/doc parsing on your input file to extract all of your documentation, comments, and
// struct field tags. It returns 2 specialized lookup maps; one for doc comments and one for the struct field tags.
// The keys to these maps are based on the names of the thing whose docs/tags you want; either "SERVICE",
// "SERVICE.FUNCTION", "MODEL", or "MODEL.FIELD".
func ParseDocumentation(ctx *Context) (Documentation, Tags, error) {
	packageDocs, err := doc.NewFromFiles(ctx.TypeInfo.Fset, ctx.TypeInfo.Syntax, ctx.Module.Name)
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

// ParseServices looks for all 'type XxxService interface' declarations and extracts all
// service/operation info from it that we need to generate our artifacts. Most of the time
// the resulting slice will only contain 1 item since its generally good design to only define
// a single service in a file, but you might have declared multiple.
func ParseServices(ctx *Context) ([]*ServiceDeclaration, error) {
	var services []*ServiceDeclaration

	for _, name := range ctx.Scope().Names() {
		interfaceType, ok := toServiceInterface(ctx.Scope().Lookup(name))
		if !ok {
			continue
		}

		service, err := ParseService(ctx, name, interfaceType)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

// ParseService accepts an interface node from the packages type tree and builds the appropriate service
// declaration containing all service and function info.
func ParseService(ctx *Context, name string, serviceInterface *types.Interface) (*ServiceDeclaration, error) {
	service := &ServiceDeclaration{
		Name:    name,
		Version: DefaultServiceVersion,
		Gateway: &GatewayServiceOptions{},
	}
	service.Gateway.Service = service

	functions, err := ParseServiceFunctions(ctx, service, serviceInterface)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", service.Name, err)
	}
	service.Functions = functions

	ApplyServiceDocumentation(ctx, service)
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
		return nil, fmt.Errorf("%s(): not a function signature type", function.Name)
	}

	// Check to make sure that we have 2 parameters and 2 return values.
	if signature.Params().Len() != 2 {
		return nil, fmt.Errorf("%s(): does not have 2 parameters", function.Name)
	}
	if signature.Results().Len() != 2 {
		return nil, fmt.Errorf("%s(): does not return 2 values", function.Name)
	}

	param1 := signature.Params().At(0)
	param2 := signature.Params().At(1)
	result1 := signature.Results().At(0)
	result2 := signature.Results().At(1)

	// Make sure that the two inputs are a context.Context and a request struct.
	if !validMethodParam1(ctx, param1) {
		return nil, fmt.Errorf("%s(): param 1 is not a context.Context", function.Name)
	}
	if !validMethodParam2(ctx, param2) {
		return nil, fmt.Errorf("%s(): param 2 is not a pointer to a request struct", function.Name)
	}

	// Make sure that the two return values are a response struct and an error
	if !validMethodReturnValue1(ctx, result1) {
		return nil, fmt.Errorf("%s(): return value 1 is not a pointer to a struct", function.Name)
	}
	if !validMethodReturnValue2(ctx, result2) {
		return nil, fmt.Errorf("%s(): return value 2 is not an error", function.Name)
	}

	// We're enforcing a convention that you define your request/response structs in the same file as the
	// services that they correspond to. Even if you want to share common types across services, that's fine,
	// but you need to define an alias or a new type where the common type is embedded in that file.
	if function.Request = ctx.ModelByName(param2.Type().String()); function.Request == nil {
		return nil, fmt.Errorf("%s(): request struct must be defined in %s", function.Name, ctx.Path)
	}
	if function.Response = ctx.ModelByName(result1.Type().String()); function.Response == nil {
		return nil, fmt.Errorf("%s(): response struct must be defined in %s", function.Name, ctx.Path)
	}

	ApplyFunctionDocumentation(ctx, function)
	return function, nil
}

// ParseModels looks for all of the structs defined in your input file and captures them as model declarations.
func ParseModels(ctx *Context) ([]*ServiceModelDeclaration, error) {
	var models []*ServiceModelDeclaration
	for _, typeName := range ctx.Scope().Names() {
		scopeObj := ctx.Scope().Lookup(typeName)
		if _, ok := toModelStruct(scopeObj); !ok {
			continue
		}

		model, err := ParseModel(ctx, scopeObj.Type())
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// ParseModel accepts an exported struct type and captures all of the model's details including its doc options
// and all of the field/type information.
func ParseModel(ctx *Context, modelType types.Type) (*ServiceModelDeclaration, error) {
	model := &ServiceModelDeclaration{
		Name: noPackage(noPointer(modelType.String())),
		Type: ParseFieldType(ctx, modelType),
	}

	fields, err := ParseModelFields(ctx, model, modelType)
	if err != nil {
		return nil, fmt.Errorf("%s: field parsing error: %v", model.Name, err)
	}

	model.Fields = fields
	return model, nil
}

// ParseModelFields accepts the info for a model struct and constructs declarations for all of the fields
// that belong to it. This includes all of the doc options and expanded type info.
func ParseModelFields(ctx *Context, model *ServiceModelDeclaration, modelType types.Type) (FieldDeclarations, error) {
	structType, ok := underlyingStruct(modelType)
	if !ok {
		return nil, fmt.Errorf("model type is not a struct")
	}

	fields := FieldDeclarations{}
	structFields := flattenedStructFields(structType)
	for _, fieldNode := range structFields {
		field, err := ParseModelField(ctx, model, fieldNode)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
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

// ParseModelField takes the type tree node for a struct field and captures the necessary info into
// a field declaration including all type info, binding info, and doc options.
func ParseModelField(ctx *Context, model *ServiceModelDeclaration, fieldVar *types.Var) (*FieldDeclaration, error) {
	field := &FieldDeclaration{
		Name:  varName(fieldVar),
		Type:  ParseFieldType(ctx, fieldVar.Type()),
		Model: model,
	}
	field.Binding = ParseBindingOptions(ctx, field, fieldVar)
	return field, nil
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
func validMethodParam1(ctx *Context, param *types.Var) bool {
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
func ApplyServiceDocumentation(ctx *Context, service *ServiceDeclaration) {
	if ctx == nil || service == nil {
		return
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

// ApplyModelDocumentation takes the documentation comment block above your struct/alias type
// declaration and applies them to the model snapshot, parsing all Doc Options in the process.
func ApplyModelDocumentation(ctx *Context, model *ServiceModelDeclaration) {
	if model == nil {
		return
	}
	model.Documentation = ctx.Documentation.ForModel(model).Trim()
}

// ApplyFieldDocumentation takes the documentation comment block above your struct field
// declaration and applies them to the model snapshot, parsing all Doc Options in the process.
func ApplyFieldDocumentation(ctx *Context, field *FieldDeclaration) {
	if field == nil {
		return
	}
	field.Documentation = ctx.Documentation.ForField(field).Trim()
}

// fieldName returns the actual field name that should be used for this attribute within a struct.
func fieldName(field *ast.Field) string {
	if embeddedField(field) {
		return noPointer(noPackage(fieldTypeName(field)))
	}
	return field.Names[0].Name
}

// embeddedField returns true if it looks as though this struct field does not have a name; it just
// has the type information.
func embeddedField(field *ast.Field) bool {
	return len(field.Names) == 0
}

// noPackage strips of any package prefixes from an identifier (e.g. "context.Context" -> "Context")
func noPackage(ident string) string {
	period := strings.LastIndex(ident, ".")
	if period < 0 {
		return ident
	}
	return ident[period+1:]
}

// noPointer strips off any "*" prefix your type identifier might have (e.g. "*Foo" -> "Foo")
func noPointer(ident string) string {
	return strings.TrimLeft(ident, "*")
}

func typeName(t types.Type) string {
	name := t.String()

	// Third party packages include the entire import path and package info (e.g. "github.com/module/pkg/subpkg.Foo")
	slash := strings.LastIndex(name, "/")
	if slash >= 0 {
		name = name[slash+1:]
	}

	// HACK: Depending on the context, types defined in the input file may be described w/ this prefix.
	// Strip that off because it's not a "real" package prefix.
	if strings.HasPrefix(name, "command-line-arguments.") {
		return name[23:]
	}
	if strings.HasPrefix(name, "*command-line-arguments.") {
		return name[24:]
	}
	return name
}

func fieldTypeName(field *ast.Field) string {
	return types.ExprString(field.Type)
}

func varName(v *types.Var) string {
	if v.Embedded() {
		return noPointer(noPackage(v.Type().String()))
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

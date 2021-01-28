package parser

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

// ParseFile parses a source code file containing a service interface declaration as well as the
// structs for the request/response inputs and outputs. It will aggregate all of the services/methods/models
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
	if ctx.Package, ctx.OutputPackage, err = ParsePackageInfo(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.Models, err = ParseServiceModels(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if ctx.Services, err = ParseServices(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if err = ApplyDocumentation(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}

	return ctx, nil
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
		Name:      packageName + "rpc",
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

// ParseServiceModels looks at all of the top-level 'type XXX' definitions looking for
// structs and type aliases that you plan to use as the requests/responses to all of
// your service functions.
func ParseServiceModels(ctx *Context) ([]*ServiceModelDeclaration, error) {
	var models []*ServiceModelDeclaration
	for _, scopeObj := range ctx.File.Scope.Objects {
		if !IsModelDeclaration(scopeObj) {
			continue
		}

		model, err := ParseServiceModel(ctx, scopeObj)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	return models, nil
}

// ParseServiceModel accepts a single 'type XXX' node and generates the model information
// that we want to capture for the service context.
func ParseServiceModel(_ *Context, modelNode *ast.Object) (*ServiceModelDeclaration, error) {
	return &ServiceModelDeclaration{
		Name: modelNode.Name,
		Node: modelNode,
	}, nil
}

// ParseServices looks for all 'type XxxService interface' declarations and extracts all
// service/operation info from it that we need to generate our artifacts. Most of the time
// the resulting slice will only contain 1 item since its generally good design to only define
// a single service in a file, but you might have declared multiple.
func ParseServices(ctx *Context) ([]*ServiceDeclaration, error) {
	var services []*ServiceDeclaration
	for _, scopeObj := range ctx.File.Scope.Objects {
		if !IsServiceDeclaration(scopeObj) {
			continue
		}

		service, err := ParseService(ctx, scopeObj)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

// ParseService accepts the syntax tree node for the 'type XxxService interface' declaration
// for a single service and extracts all meaningful information about it. The service/function/model
// info is packaged up in a service declaration which you can add to your Context.
func ParseService(ctx *Context, serviceObj *ast.Object) (*ServiceDeclaration, error) {
	service := &ServiceDeclaration{
		Name: serviceObj.Name,
		Node: serviceObj,
	}

	methods, err := ParseServiceMethods(ctx, serviceObj)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", serviceObj.Name, err)
	}
	service.Methods = methods
	return service, nil
}

// ParseServiceMethods accepts the syntax tree node for a 'type XxxService interface' declaration,
// iterates the functions it defines and creates declarations for each one with just the info from
// it that we need when building clients/gateways for the service.
func ParseServiceMethods(ctx *Context, serviceNode *ast.Object) ([]*ServiceMethodDeclaration, error) {
	interfaceType, _ := serviceNode.
		Decl.(*ast.TypeSpec).
		Type.(*ast.InterfaceType)

	var methods []*ServiceMethodDeclaration
	for _, methodObj := range interfaceType.Methods.List {
		method, err := ParseServiceMethod(ctx, serviceNode, methodObj)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", serviceNode.Name, err)
		}
		methods = append(methods, method)
	}
	return methods, nil
}

// ParseServiceMethod accepts the ast node for a service interface and one if its functions, then
// aggregates all of the information about that service operation for the context. It captures
// the name as well as HTTP-related info (status, method, path) to use in the gateway.
func ParseServiceMethod(ctx *Context, serviceNode *ast.Object, methodObj *ast.Field) (*ServiceMethodDeclaration, error) {
	name := fieldName(methodObj)
	method := &ServiceMethodDeclaration{
		Name:       name,
		Node:       methodObj,
		HTTPStatus: http.StatusOK,
		HTTPMethod: "POST",
		HTTPPath:   "/" + serviceNode.Name + "." + fieldName(methodObj),
	}

	function := methodObj.Type.(*ast.FuncType)

	// Check to make sure that we have 2 parameters w/ the correct types (context and your request)
	if len(function.Params.List) != 2 {
		return nil, fmt.Errorf("%s: does not have 2 parameters", name)
	}
	if !isValidParam1(ctx, function.Params.List[0]) {
		return nil, fmt.Errorf("%s: first param is not a context.Context", name)
	}
	if !isValidParam2(ctx, function.Params.List[1]) {
		return nil, fmt.Errorf("%s: second param type is defined in this file", name)
	}

	// Check to make sure that we have 2 return values (your response type and an error)
	if len(function.Results.List) != 2 {
		return nil, fmt.Errorf("%s: does not return 2 values", name)
	}
	if !isValidReturnValue1(ctx, function.Results.List[0]) {
		return nil, fmt.Errorf("%s: first return value is not deifned in this file", name)
	}
	if !isValidReturnValue2(ctx, function.Results.List[1]) {
		return nil, fmt.Errorf("%s: second return value is not an error", name)
	}

	// Connect the model/struct declarations for request/response to this method.
	method.Request = ctx.ModelByName(typeName(function.Params.List[1]))
	method.Response = ctx.ModelByName(typeName(function.Results.List[0]))

	return method, nil
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

// The first param to all service methods should be a standard "context.Context"
func isValidParam1(_ *Context, param *ast.Field) bool {
	// I know... doesn't handle if you decided to alias the context class, but that
	// should be a fairly infrequent case. I'm going to wait for feedback to determine
	// if there's a valid need to compare the X in the selector expression to your
	// file's imports and see if they resolve to the right context class.
	return typeName(param) == "context.Context"
}

// The parameter should be a "request" struct/type that is defined in the file we're parsing,
// so to be valid, we must have parsed a model with the same type name earlier.
func isValidParam2(ctx *Context, param *ast.Field) bool {
	return ctx.ModelByName(typeName(param)) != nil
}

// The first return value should be a "response" struct/type that you also defined in the file
// that we're parsing. There needs to be a struct/type of the same name that we parsed earlier.
func isValidReturnValue1(ctx *Context, returnVal *ast.Field) bool {
	return ctx.ModelByName(typeName(returnVal)) != nil
}

// The second return value should be an error for idiomatic failure handling.
func isValidReturnValue2(_ *Context, returnVal *ast.Field) bool {
	return typeName(returnVal) == "error"
}

// IsServiceDeclaration analyzes a node from the AST and determines if it's a `type XxxService interface`
// declaration defining one of your services. In addition to being an interface it must also
// be exported (e.g. "FooService" instead of "fooService") as well as follow the naming convention of
// ending with the word "Service" (e.g. "FooService" instead of just "Foo").
func IsServiceDeclaration(astObj *ast.Object) bool {
	// Only looking for 'type' declarations (e.g. 'type XXX interface')
	if astObj.Kind != ast.Typ {
		return false
	}
	typeSpec, ok := astObj.Decl.(*ast.TypeSpec)
	if !ok {
		return false
	}

	// For RPC purposes, we only expose exported interfaces.
	if !typeSpec.Name.IsExported() {
		return false
	}

	// Your service declaration must be an interface.
	_, ok = typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return false
	}

	// We're enforcing the convention that the "foo" service is called "FooService"
	return strings.HasSuffix(astObj.Name, "Service")
}

// IsModelDeclaration looks at a node from your file's AST and returns true if it's a type
// declaration that can be used as one of our request/response values to our service operations.
func IsModelDeclaration(astObj *ast.Object) bool {
	// Only looking for 'type' declarations (e.g. 'type XXX struct')
	if astObj.Kind != ast.Typ {
		return false
	}
	typeSpec, ok := astObj.Decl.(*ast.TypeSpec)
	if !ok {
		return false
	}

	// Since we're using the same types in the auto-generated clients, all request/response
	// models must be exported so code from other packages/services can access them.
	if !typeSpec.Name.IsExported() {
		return false
	}

	// The model type must either be a struct or some sort of type alias.
	switch typeSpec.Type.(type) {
	case *ast.StructType:
		return true
	case *ast.Ident: // type alias to another type in this package (e.g. "type Foo Bar")
		return true
	case *ast.SelectorExpr: // type alias to a type in another package (e.g. "type Foo other.Bar")
		return true
	default:
		return false
	}
}

// ApplyDocumentation runs GoDoc parsing on your context's file and adds all of your source's documentation
// comments to the services/methods/models in the context. This *does* mutate the values on the context.
// In addition to regurgitating the comments, this will ultimately parse all of the Doc Options
// that might appear in the comments.
func ApplyDocumentation(ctx *Context) error {
	docs, err := doc.NewFromFiles(ctx.FileSet, []*ast.File{ctx.File}, ctx.Module.Name)
	if err != nil {
		return err
	}

	// Look through all of the top-level type definitions for structs/aliases you used as
	// request/response models and process their comments.
	for _, typeDef := range docs.Types {
		model := ctx.ModelByName(typeDef.Name)
		ApplyModelDocumentation(ctx, model, typeDef.Doc)
	}

	// Look through all of the top-level service interface definitions and apply all of the
	// documentation options/comments to the service and its methods.
	for _, typeDef := range docs.Types {
		service := ctx.ServiceByName(typeDef.Name)
		if service == nil {
			continue
		}
		ApplyServiceDocumentation(ctx, service, typeDef.Doc)

		// You might ask yourself why we're going back to the original syntax tree to iterate
		// the service methods rather than iterating 'typeDef.Funcs'. Well... because in all of
		// my testing this stuff out on real .go files, both ".Methods" and ".Funcs" are nil
		// on the service interface documentation nodes. Even when the functions have GoDoc
		// comments, they're nil.
		//
		// I'm probably doing something wrong to get in this situation or maybe I just don't fully
		// understand the syntax tree parsing logic well enough (likely both). But here's what I'm
		// observing: all of the documentation/comment data on the original AST is missing for
		// the top-level type definitions (services and models), so I need to actually invoke the
		// GoDoc parser (doc.NewFromFiles() above) to get those comments. For the interface functions,
		// however, I'm seeing the exact opposite. The original AST *does* have the comments on those
		// interface functions, but I can't seem to get to them when using the 'docs' tree.
		//
		// That's why I'm mixing and matching where I'm getting the docs from. Models/services come
		// from the GoDoc parser and the function docs come from the original AST nodes. Maybe one day
		// I'll learn what the heck is going on and deal with it properly, but for now this does
		// effectively give me what I want - the complete doc comments for all items in my context.
		for _, methodObj := range service.InterfaceNode().Methods.List {
			if methodObj.Doc == nil {
				continue
			}
			method := service.MethodByName(fieldName(methodObj))
			ApplyMethodDocumentation(ctx, method, methodObj.Doc.Text())
		}
	}
	return nil
}

// ApplyServiceDocumentation takes the documentation comment block above your interface type
// declaration and applies them to the service snapshot, parsing all Doc Options in the process.
func ApplyServiceDocumentation(_ *Context, service *ServiceDeclaration, comments string) {
	if service == nil {
		return
	}
	if comments == "" {
		return
	}
	for _, line := range strings.Split(comments, "\n") {
		switch {
		case strings.HasPrefix(line, "PATH "):
			service.HTTPPathPrefix = normalizePathSegment(line[5:])
		case strings.HasPrefix(line, "PREFIX "):
			service.HTTPPathPrefix = normalizePathSegment(line[7:])
		default:
			service.Documentation = append(service.Documentation, line)
		}
	}
	service.Documentation = service.Documentation.Trim()
}

// ApplyMethodDocumentation takes the documentation comment block above your interface function
// declaration and applies them to the method snapshot, parsing all Doc Options in the process.
func ApplyMethodDocumentation(_ *Context, method *ServiceMethodDeclaration, comments string) {
	if method == nil {
		return
	}
	if comments == "" {
		return
	}
	for _, line := range strings.Split(comments, "\n") {
		switch {
		case strings.HasPrefix(line, "GET /"):
			method.HTTPMethod = http.MethodGet
			method.HTTPPath = normalizePathSegment(line[4:])
		case strings.HasPrefix(line, "PUT /"):
			method.HTTPMethod = http.MethodPut
			method.HTTPPath = normalizePathSegment(line[4:])
		case strings.HasPrefix(line, "POST /"):
			method.HTTPMethod = http.MethodPost
			method.HTTPPath = normalizePathSegment(line[5:])
		case strings.HasPrefix(line, "PATCH /"):
			method.HTTPMethod = http.MethodPatch
			method.HTTPPath = normalizePathSegment(line[6:])
		case strings.HasPrefix(line, "DELETE /"):
			method.HTTPMethod = http.MethodDelete
			method.HTTPPath = normalizePathSegment(line[7:])
		case strings.HasPrefix(line, "OPTIONS /"):
			method.HTTPMethod = http.MethodOptions
			method.HTTPPath = normalizePathSegment(line[8:])
		case strings.HasPrefix(line, "HEAD /"):
			method.HTTPMethod = http.MethodHead
			method.HTTPPath = normalizePathSegment(line[5:])
		case strings.HasPrefix(line, "HTTP "):
			method.HTTPStatus = parseHTTPStatus(line[5:])
		default:
			method.Documentation = append(method.Documentation, line)
		}
	}
	method.Documentation = method.Documentation.Trim()
}

// ApplyModelDocumentation takes the documentation comment block above your struct/alias type
// declaration and applies them to the model snapshot, parsing all Doc Options in the process.
func ApplyModelDocumentation(_ *Context, model *ServiceModelDeclaration, comments string) {
	if model == nil {
		return
	}
	if comments == "" {
		return
	}
	model.Documentation = strings.Split(comments, "\n")
	model.Documentation = model.Documentation.Trim()
}

func fieldName(field *ast.Field) string {
	if len(field.Names) == 0 {
		return ""
	}
	return field.Names[0].Name
}

func typeName(field *ast.Field) string {
	switch fieldType := field.Type.(type) {
	case *ast.Ident:
		return fieldType.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%v.%v", fieldType.X, fieldType.Sel)
	case *ast.StarExpr:
		return fmt.Sprintf("%v", fieldType.X)
	default:
		return ""
	}
}

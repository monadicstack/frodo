package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

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
		File:         file,
		Path:         inputPath,
		AbsolutePath: absolutePath,
	}

	if err = ParseModuleInfo(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if err = ParsePackageInfo(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if err = ParseModelTypes(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	if err = ParseServiceInterfaces(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputPath, err)
	}
	return ctx, nil
}

func ParsePackageInfo(ctx *Context) error {
	packageName := ctx.File.Name.Name
	packageDir := filepath.Dir(ctx.Path)
	packageDirRelative := strings.TrimPrefix(packageDir, ctx.Module.Directory)

	ctx.Package = &PackageDeclaration{
		Name:      packageName,
		Import:    filepath.Join(ctx.Module.Name, packageDirRelative),
		Directory: filepath.Dir(ctx.Path),
	}
	ctx.OutputPackage = &PackageDeclaration{
		Name:      packageName + "rpc",
		Import:    filepath.Join(ctx.Package.Import, "gen"),
		Directory: filepath.Join(ctx.Package.Directory, "gen"),
	}
	return nil
}

func ParseModuleInfo(ctx *Context) error {
	inputFilePath, err := filepath.Abs(ctx.Path)
	if err != nil {
		return fmt.Errorf("unable to determine absolute path: %w", err)
	}

	// Look in the input file's directory (an all of its parents/ancestors) for the "go.mod" file.
	goModPath, err := findGoDotMod(filepath.Dir(inputFilePath))
	if err != nil {
		return err
	}
	// Read/parse the "go.mod" file so we can extract the module/package info we need.
	goModData, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return err
	}
	goModFile, err := modfile.Parse(goModPath, goModData, nil)
	if err != nil {
		return err
	}

	ctx.Module = &ModuleDeclaration{
		Name:      goModFile.Module.Mod.Path,
		Directory: filepath.Dir(goModPath),
	}
	return nil
}

func findGoDotMod(dirName string) (string, error) {
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
	return findGoDotMod(filepath.Dir(dirName))
}

func ParseModelTypes(ctx *Context) error {
	for _, scopeObj := range ctx.File.Scope.Objects {
		if !IsModelDeclaration(scopeObj) {
			continue
		}

		err := ParseModelStruct(ctx, scopeObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseModelStruct(ctx *Context, astObject *ast.Object) error {
	ctx.AddModel(&ServiceModelDeclaration{
		Name: astObject.Name,
		Node: astObject,
	})
	return nil
}

func ParseServiceInterfaces(ctx *Context) error {
	for _, scopeObj := range ctx.File.Scope.Objects {
		if !IsServiceDeclaration(scopeObj) {
			continue
		}

		err := ParseService(ctx, scopeObj)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseService(ctx *Context, serviceObj *ast.Object) error {
	service := &ServiceDeclaration{
		Name: serviceObj.Name,
		Node: serviceObj,
	}

	ctx.currentService = service
	defer func() {
		ctx.currentService = nil
	}()

	interfaceType, _ := serviceObj.
		Decl.(*ast.TypeSpec).
		Type.(*ast.InterfaceType)

	for _, methodObj := range interfaceType.Methods.List {
		err := ParseServiceMethod(ctx, methodObj)
		if err != nil {
			return fmt.Errorf("%s: %w", serviceObj.Name, err)
		}
	}

	ctx.AddService(service)
	return nil
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

func ParseServiceMethod(ctx *Context, methodObj *ast.Field) error {
	name := fieldName(methodObj)
	method := &ServiceMethodDeclaration{
		Name:       name,
		Node:       methodObj,
		HTTPStatus: http.StatusOK,
		HTTPMethod: "POST",
		HTTPPath:   "/" + ctx.currentService.Name + "." + fieldName(methodObj),
	}

	ctx.currentMethod = method
	defer func() {
		ctx.currentMethod = nil
	}()

	function := methodObj.Type.(*ast.FuncType)

	// Check to make sure that we have 2 parameters w/ the correct types (context and your request)
	if len(function.Params.List) != 2 {
		return fmt.Errorf("%s: does not have 2 parameters", name)
	}
	if !isValidParam1(ctx, function.Params.List[0]) {
		return fmt.Errorf("%s: first param is not a context.Context", name)
	}
	if !isValidParam2(ctx, function.Params.List[1]) {
		return fmt.Errorf("%s: second param type is defined in this file", name)
	}

	// Check to make sure that we have 2 return values (your response type and an error)
	if len(function.Results.List) != 2 {
		return fmt.Errorf("%s: does not return 2 values", name)
	}
	if !isValidReturnValue1(ctx, function.Results.List[0]) {
		return fmt.Errorf("%s: first return value is not deifned in this file", name)
	}
	if !isValidReturnValue2(ctx, function.Results.List[1]) {
		return fmt.Errorf("%s: second return value is not an error", name)
	}

	// Connect the model/struct declarations for request/response to this method.
	method.Request = ctx.ModelByName(typeName(function.Params.List[1]))
	method.Response = ctx.ModelByName(typeName(function.Results.List[0]))

	// Check the doc comments for the function to determine if they're providing
	// a custom method/path for the endpoint as opposed to the default RPC-style we assign.
	applyDocCommentOptions(ctx, methodObj, method)

	ctx.currentService.AddMethod(method)
	return nil
}

func applyDocCommentOptions(_ *Context, methodObj *ast.Field, method *ServiceMethodDeclaration) {
	if methodObj.Doc == nil {
		return
	}
	for _, doc := range methodObj.Doc.List {
		comment := doc.Text
		comment = strings.TrimSpace(comment)
		comment = strings.TrimPrefix(comment, "//")
		comment = strings.TrimSpace(comment)

		switch {
		case strings.HasPrefix(comment, "GET /"):
			method.HTTPMethod = http.MethodGet
			method.HTTPPath = comment[4:]
		case strings.HasPrefix(comment, "PUT /"):
			method.HTTPMethod = http.MethodPut
			method.HTTPPath = comment[4:]
		case strings.HasPrefix(comment, "POST /"):
			method.HTTPMethod = http.MethodPost
			method.HTTPPath = comment[5:]
		case strings.HasPrefix(comment, "PATCH /"):
			method.HTTPMethod = http.MethodPatch
			method.HTTPPath = comment[6:]
		case strings.HasPrefix(comment, "DELETE /"):
			method.HTTPMethod = http.MethodDelete
			method.HTTPPath = comment[7:]
		case strings.HasPrefix(comment, "OPTIONS /"):
			method.HTTPMethod = http.MethodOptions
			method.HTTPPath = comment[8:]
		case strings.HasPrefix(comment, "HEAD /"):
			method.HTTPMethod = http.MethodHead
			method.HTTPPath = comment[5:]
		case strings.HasPrefix(comment, "HTTP "):
			method.HTTPStatus = parseHTTPStatus(comment[5:])
		}
	}
}

func parseHTTPStatus(statusText string) int {
	status, err := strconv.ParseInt(statusText, 10, 64)
	if err != nil {
		return http.StatusOK
	}
	return int(status)
}

// The first param to all service methods should be a standard "context.Context"
func isValidParam1(_ *Context, param *ast.Field) bool {
	return typeName(param) == "context.Context"
}

// The parameter should be a "request" struct that is defined in the file we're parsing, so
// to be valid, we must have parsed a model with the same type name earlier.
func isValidParam2(ctx *Context, param *ast.Field) bool {
	return ctx.ModelByName(typeName(param)) != nil
}

// The first return value should be a "response" struct that you also defined in the file
// that we're parsing. There needs to be a struct/type of the same name that we parsed earlier.
func isValidReturnValue1(ctx *Context, returnVal *ast.Field) bool {
	return ctx.ModelByName(typeName(returnVal)) != nil
}

// The second return value should be an error for idiomatic failure handling.
func isValidReturnValue2(_ *Context, returnVal *ast.Field) bool {
	return typeName(returnVal) == "error"
}

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

	// We're enforcing the convention that the "foo" service is called "FooService", etc.
	return strings.HasSuffix(astObj.Name, "Service")
}

func IsModelDeclaration(astObj *ast.Object) bool {
	// Only looking for 'type' declarations (e.g. 'type XXX interface')
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

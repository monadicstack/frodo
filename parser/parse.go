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
	if ctx.Package, ctx.OutputPackage, err = ParsePackageInfo(ctx); err != nil {
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
	return ctx, nil
}

// ParseTypeInformation runs the syntax tree through the "go/types" processor so that we get detailed
// type information on all of the structs/types we defined, their fields, and the parameters/outputs
// of our service functions.
func ParseTypeInformation(ctx *Context) (*packages.Package, error) {
	config := &packages.Config{
		Tests: false,
		Mode:  packages.NeedDeps | packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
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

// lookupFieldList accepts the AST node for a type definition and returns a slice of all
// fields/attributes that belong to this type.
//
// IMPORTANT! The resulting slice will contain the attributes of all embedded types as well.
// As a result, the slice should contain explicitly defined attributes as well as those it
// sugar-coats due to embedding.
func lookupFieldList(modelNode *ast.Object) ([]*ast.Field, error) {
	if modelNode == nil {
		return nil, nil
	}
	typeSpec, ok := modelNode.Decl.(*ast.TypeSpec)
	if !ok {
		return nil, fmt.Errorf("unable to look up fields for non-type spec")
	}
	return lookupFieldListForType(typeSpec.Type)
}

// lookupFieldListForType behaves the same as lookupFieldList, but works by taking an AST type
// expression that you already have in-hand.
func lookupFieldListForType(typeExpr ast.Expr) ([]*ast.Field, error) {
	switch t := typeExpr.(type) {
	case *ast.StructType:
		return flattenEmbeddedFields(t.Fields.List)
	case *ast.Ident:
		return lookupFieldList(t.Obj)
	case *ast.SelectorExpr:
		return lookupFieldList(t.Sel.Obj)
	default:
		return nil, nil
	}
}

// flattenEmbeddedFields takes the exact list of fields defined on a struct and expands the list
// to include any "inherited" fields that came from fields that were actually embedded types.
func flattenEmbeddedFields(fields []*ast.Field) ([]*ast.Field, error) {
	var results []*ast.Field
	for _, field := range fields {
		if !embeddedField(field) {
			results = append(results, field)
			continue
		}

		embeddedFields, err := lookupFieldListForType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("embedded field lookup error: %s: %v", fieldName(field), err)
		}
		results = append(results, embeddedFields...)
	}
	return results, nil
}

// ParseFieldType looks at the Go parser's type information for a given model attribute and extracts
// all of the various info we need to get a complete picture of the type and how to unravel any
// aliasing that might be going on.
func ParseFieldType(ctx *Context, t types.Type) *FieldType {
	fieldType := &FieldType{
		Name:       t.String(),
		Type:       t,
		Underlying: underlyingType(t),
	}

	switch underlying := fieldType.Underlying.(type) {
	case *types.Pointer:
		fieldType.Pointer = true
		fieldType.Type = underlying.Elem()
		fieldType.Underlying = underlyingType(fieldType.Type)
	case *types.Array:
		fieldType.Elem = ParseFieldType(ctx, underlying.Elem())
	case *types.Slice:
		fieldType.Elem = ParseFieldType(ctx, underlying.Elem())
	case *types.Chan:
		fieldType.Elem = ParseFieldType(ctx, underlying.Elem())
	case *types.Map:
		fieldType.Key = ParseFieldType(ctx, underlying.Key())
		fieldType.Elem = ParseFieldType(ctx, underlying.Elem())
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
	case *types.Chan:
		return "object"
	case *types.Map:
		return "object"
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

func ParseDocumentation(ctx *Context) (Documentation, Tags, error) {
	packageDocs, err := doc.NewFromFiles(ctx.TypeInfo.Fset, ctx.TypeInfo.Syntax, ctx.Module.Name)
	if err != nil {
		return nil, nil, err
	}

	docs := Documentation{}
	tags := Tags{}

	// First iterate through the service interfaces and capture the interface/function docs.
	for _, serviceType := range packageDocs.Types {
		interfaceNode, ok := toInterfaceTypeNode(serviceType)
		if !ok {
			continue
		}
		docs[serviceType.Name] = strings.TrimSpace(serviceType.Doc)

		for _, functionNode := range interfaceNode.Methods.List {
			if functionNode.Doc == nil {
				continue
			}
			docs[serviceType.Name+"."+fieldName(functionNode)] = functionNode.Doc.Text()
		}
	}

	// Now iterate the models to capture the model/field docs.
	for _, modelType := range packageDocs.Types {
		structNode, ok := toStructTypeNode(modelType)
		if !ok {
			continue
		}
		docs[modelType.Name] = strings.TrimSpace(modelType.Doc)

		for _, fieldNode := range structNode.Fields.List {
			if fieldNode.Doc == nil {
				continue
			}

			lookupKey := modelType.Name + "." + fieldName(fieldNode)
			docs[lookupKey] = fieldNode.Doc.Text()

			if fieldNode.Tag != nil {
				tags[lookupKey] = fieldNode.Tag.Value
			}
		}
	}

	return docs, tags, nil
}

func toStructTypeNode(t *doc.Type) (*ast.StructType, bool) {
	typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil, false
	}
	structNode, ok := typeSpec.Type.(*ast.StructType)
	return structNode, ok
}

func toInterfaceTypeNode(t *doc.Type) (*ast.InterfaceType, bool) {
	typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil, false
	}
	interfaceNode, ok := typeSpec.Type.(*ast.InterfaceType)
	return interfaceNode, ok
}

func ToServiceInterface(typeObj types.Object) (*types.Interface, bool) {
	// Enforce the naming convention that services end w/ the word "Service"
	if !strings.HasSuffix(typeObj.Name(), "Service") {
		return nil, false
	}
	return underlyingInterface(typeObj.Type())
}

func ToModelStruct(typeObj types.Object) (*types.Struct, bool) {
	return underlyingStruct(typeObj.Type())
}

// ParseServices looks for all 'type XxxService interface' declarations and extracts all
// service/operation info from it that we need to generate our artifacts. Most of the time
// the resulting slice will only contain 1 item since its generally good design to only define
// a single service in a file, but you might have declared multiple.
func ParseServices(ctx *Context) ([]*ServiceDeclaration, error) {
	var services []*ServiceDeclaration

	for _, typeName := range ctx.Scope().Names() {
		interfaceType, ok := ToServiceInterface(ctx.Scope().Lookup(typeName))
		if !ok {
			continue
		}

		service, err := ParseService(ctx, typeName, interfaceType)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, nil
}

func ParseService(ctx *Context, name string, serviceInterface *types.Interface) (*ServiceDeclaration, error) {
	service := &ServiceDeclaration{
		Name:    name,
		Version: "0.1.0",
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
		return nil, fmt.Errorf("%s: not a function signature type", function.Name)
	}

	// Check to make sure that we have 2 parameters w/ the correct types (context and your request)
	if signature.Params().Len() != 2 {
		return nil, fmt.Errorf("%s: does not have 2 parameters", function.Name)
	}
	if !validMethodParam1(ctx, signature.Params().At(0)) {
		return nil, fmt.Errorf("%s: param 1 is not a context.Context", function.Name)
	}
	if !validMethodParam2(ctx, signature.Params().At(1)) {
		return nil, fmt.Errorf("%s: param 2 is not a pointer to a request struct", function.Name)
	}

	// Check to make sure that we have 2 return values (your response type and an error)
	if signature.Results().Len() != 2 {
		return nil, fmt.Errorf("%s: does not return 2 values", function.Name)
	}
	if !validMethodReturnValue1(ctx, signature.Results().At(0)) {
		return nil, fmt.Errorf("%s: return value 1 is not a pointer to a struct", function.Name)
	}
	if !validMethodReturnValue2(ctx, signature.Results().At(1)) {
		return nil, fmt.Errorf("%s: return value 2 is not an error", function.Name)
	}

	function.Request = ctx.ModelByName(signature.Params().At(1).Type().String())
	function.Response = ctx.ModelByName(signature.Results().At(0).Type().String())

	ApplyFunctionDocumentation(ctx, function)
	return function, nil
}

func ParseModels(ctx *Context) ([]*ServiceModelDeclaration, error) {
	var models []*ServiceModelDeclaration
	for _, typeName := range ctx.Scope().Names() {
		scopeObj := ctx.Scope().Lookup(typeName)
		if !scopeObj.Exported() {
			continue
		}
		if _, ok := ToModelStruct(scopeObj); !ok {
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

func ParseModelFields(ctx *Context, model *ServiceModelDeclaration, modelType types.Type) (FieldDeclarations, error) {
	fields := FieldDeclarations{}

	structType, ok := underlyingStruct(modelType)
	if !ok {
		return nil, fmt.Errorf("model type is not a struct")
	}

	for i := 0; i < structType.NumFields(); i++ {
		fieldNode := structType.Field(i)
		if !fieldNode.Exported() {
			continue
		}

		field, err := ParseModelField(ctx, model, fieldNode)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

func ParseModelField(ctx *Context, model *ServiceModelDeclaration, fieldVar *types.Var) (*FieldDeclaration, error) {
	field := &FieldDeclaration{
		Name:  varName(fieldVar),
		Type:  ParseFieldType(ctx, fieldVar.Type()),
		Model: model,
	}
	field.Binding = ParseBindingOptions(ctx, field, fieldVar)
	return field, nil
}

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

func validMethodParam2(_ *Context, param *types.Var) bool {
	if _, ok := underlyingPointer(param.Type()); !ok {
		return false
	}
	if _, ok := underlyingStruct(param.Type()); !ok {
		return false
	}
	return true
}

func validMethodReturnValue1(ctx *Context, param *types.Var) bool {
	// It has the same semantics - must be a pointer to a struct.
	return validMethodParam2(ctx, param)
}

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
			service.Gateway.PathPrefix = normalizePathSegment(line[5:])
		case strings.HasPrefix(line, "PREFIX "):
			service.Gateway.PathPrefix = normalizePathSegment(line[7:])
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
		case strings.HasPrefix(line, "GET /"):
			function.Gateway.Method = http.MethodGet
			function.Gateway.Path = normalizePathSegment(line[4:])
		case strings.HasPrefix(line, "PUT /"):
			function.Gateway.Method = http.MethodPut
			function.Gateway.Path = normalizePathSegment(line[4:])
		case strings.HasPrefix(line, "POST /"):
			function.Gateway.Method = http.MethodPost
			function.Gateway.Path = normalizePathSegment(line[5:])
		case strings.HasPrefix(line, "PATCH /"):
			function.Gateway.Method = http.MethodPatch
			function.Gateway.Path = normalizePathSegment(line[6:])
		case strings.HasPrefix(line, "DELETE /"):
			function.Gateway.Method = http.MethodDelete
			function.Gateway.Path = normalizePathSegment(line[7:])
		case strings.HasPrefix(line, "HEAD /"):
			function.Gateway.Method = http.MethodHead
			function.Gateway.Path = normalizePathSegment(line[5:])
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
		return noPointer(noPackage(typeName(field)))
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

func typeName(field *ast.Field) string {
	return types.ExprString(field.Type)
}

func varName(v *types.Var) string {
	if v.Embedded() {
		return noPointer(noPackage(v.Type().String()))
	}
	return v.Name()
}

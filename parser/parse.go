package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func ParseFile(inputFile string) (*Context, error) {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, inputFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("[%s] unable to parse go file: %w", inputFile, err)
	}

	ctx := &Context{
		File:    file,
		Path:    inputFile,
		Package: file.Name.Name,
	}

	// Step 1: Identify all of the struct/type declarations for the request/response models.
	if err = ParseModelTypes(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputFile, err)
	}
	// Step 2: Now that we know the possible inputs/outputs, parse the service interfaces and their functions
	if err = ParseServiceInterfaces(ctx); err != nil {
		return nil, fmt.Errorf("[%s] parse error: %w", inputFile, err)
	}
	return ctx, nil
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
	fmt.Printf(">>>>>>>>> TYPE NAME: %T\n", field.Type)
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
		Name:          name,
		Node:          methodObj,
		GatewayMethod: "POST",
		GatewayPath:   ctx.currentService.Name + "." + name,
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

	method.Request = ctx.ModelByName(typeName(function.Params.List[1]))
	method.Response = ctx.ModelByName(typeName(function.Results.List[0]))
	ctx.currentService.AddMethod(method)
	return nil
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

type Context struct {
	File     *ast.File
	Path     string
	Package  string
	Services []*ServiceDeclaration
	Models   []*ServiceModelDeclaration

	currentService *ServiceDeclaration
	currentMethod  *ServiceMethodDeclaration
}

func (ctx *Context) AddService(service *ServiceDeclaration) {
	ctx.Services = append(ctx.Services, service)
}

func (ctx *Context) AddModel(model *ServiceModelDeclaration) {
	ctx.Models = append(ctx.Models, model)
}

func (ctx Context) ModelByName(name string) *ServiceModelDeclaration {
	fmt.Println(">>>> Looking for ", name)
	for _, model := range ctx.Models {
		if model.Name == name {
			return model
		}
	}
	return nil
}

type ServiceDeclaration struct {
	Name    string
	Methods []*ServiceMethodDeclaration
	Node    *ast.Object
}

func (service *ServiceDeclaration) AddMethod(method *ServiceMethodDeclaration) {
	service.Methods = append(service.Methods, method)
}

func (service ServiceDeclaration) String() string {
	buf := strings.Builder{}
	buf.WriteString(service.Name)
	for _, method := range service.Methods {
		buf.WriteString("\n")
		buf.WriteString("    ." + method.String())
	}
	return buf.String()
}

func (service ServiceDeclaration) MethodByName(name string) *ServiceMethodDeclaration {
	for _, m := range service.Methods {
		if m.Name == name {
			return m
		}
	}
	return nil
}

type ServiceMethodDeclaration struct {
	Name          string
	Request       *ServiceModelDeclaration
	Response      *ServiceModelDeclaration
	GatewayMethod string
	GatewayPath   string
	Node          *ast.Field
}

func (method ServiceMethodDeclaration) String() string {
	return fmt.Sprintf("%s(context.Context, %v) (%v, error)",
		method.Name,
		method.Request,
		method.Response,
	)
}

type ServiceModelDeclaration struct {
	Name string
	Node *ast.Object
}

func (model ServiceModelDeclaration) String() string {
	return model.Name
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

func IsPackageDeclaration(astObj *ast.Object) bool {
	return astObj.Kind == ast.Pkg
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

	// Your input must be a struct
	// TODO: support type aliases so you can re-use common inputs/outputs
	_, ok = typeSpec.Type.(*ast.StructType)
	if !ok {
		return false
	}

	return true

	// We're enforcing the convention that the input to the method "CreateFoo" must be "CreateFooRequest"
	// and the output must be called "CreateFooResponse"
	//return strings.HasSuffix(astObj.Name, "Request") || strings.HasSuffix(astObj.Name, "Response")
}

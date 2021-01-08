package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/robsignorelli/expose/generate"
	"github.com/robsignorelli/expose/parser"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatalf("Usage: goson [go-file]")
	}

	inputFileName := os.Args[1]
	log.Printf("[exposec] Parsing service definitions: %s", inputFileName)

	ctx, err := parser.ParseFile(inputFileName)
	crapPants(err)

	generateServer(ctx, inputFileName)
	generateClient(ctx, inputFileName)
}

func generateServer(ctx *parser.Context, inputFileName string) {
	outputFileName := strings.TrimSuffix(inputFileName, ".go") + ".gen.server.go"
	log.Printf("[exposec] Writing gateway: %s -> %s", inputFileName, outputFileName)

	_ = os.Remove(outputFileName)
	outputFile, err := os.Create(outputFileName)
	crapPants(err)
	defer outputFile.Close()

	err = generate.Server(ctx, outputFile)
	crapPants(err)
}

func generateClient(ctx *parser.Context, inputFileName string) {
	outputFileName := strings.TrimSuffix(inputFileName, ".go") + ".gen.client.go"
	log.Printf("[exposec] Writing client: %s -> %s", inputFileName, outputFileName)

	_ = os.Remove(outputFileName)
	outputFile, err := os.Create(outputFileName)
	crapPants(err)
	defer outputFile.Close()

	err = generate.Client(ctx, outputFile)
	crapPants(err)
}

func crapPants(err error) {
	if err != nil {
		fmt.Printf("[exposec] fatal error: %v\n", err)
		os.Exit(1)
	}
}

/*
func ParseFile(fileSet *token.FileSet, path string) *DeclarationContext {
	code, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("unable to parse go file [%s]: %v", path, err)
		return nil
	}

	ctx := &DeclarationContext{
		SourcePath: path,
	}

	// Step 1: Identify/parse all of the struct declarations for the request/response models.
	for _, scopeObj := range code.Scope.Objects {
		if IsModelDeclaration(scopeObj) {
			model := ParseModelStruct(scopeObj)
			ctx.Models = append(ctx.Models, model)
		}
	}

	// Step 2: Parse all of the service interfaces and their functions (now that we know the inputs/outputs).
	for _, scopeObj := range code.Scope.Objects {
		if IsServiceDeclaration(scopeObj) {
			service := ParseService(scopeObj)
			ctx.Services = append(ctx.Services, service)
		}
	}

	return ctx
}

func ParseService(ctx *DeclarationContext, serviceObj *ast.Object) (*ServiceDeclaration, error) {
	declaration := &ServiceDeclaration{
		Name: serviceObj.Name,
		Node: serviceObj,
	}

	interfaceType, _ := serviceObj.
		Decl.(*ast.TypeSpec).
		Type.(*ast.InterfaceType)

	for _, methodObj := range interfaceType.Methods.List {
		methodDeclaration, err := ParseServiceMethod(serviceObj, methodObj)
		if err != nil {
			return nil, fmt.Errorf("invalid method: %s.%s: %v", serviceObj.Name, methodObj.Names[0].Name, err)
		}

		declaration.Methods = append(declaration.Methods, methodDeclaration)
	}
	return declaration, nil
}

func fieldName(obj *ast.Field) string {
	if obj == nil {
		return ""
	}
	if len(obj.Names) == 0 {
		return ""
	}
	return obj.Names[0].Name
}

func ParseServiceMethod(serviceObj *ast.Object, methodObj *ast.Field) (*ServiceMethodDeclaration, error) {
	funcType := methodObj.Type.(*ast.FuncType)
	if hasValidParams(funcType) {
		return nil, fmt.Errorf("invalid service method: %s.%s: must have two arguments; context.Context and your request type",
			serviceObj.Name,
			fieldName(methodObj),
		)
	}

	fmt.Println(">>>>>> METHOD:", methodObj, methodObj.Names[0].Name)
	fmt.Printf(">>>>>>>>>>> : %T\n", methodObj.Type)
	declaration := &ServiceMethodDeclaration{Name: methodObj.Names[0].Name}
	return declaration, nil
}

func hasValidParams(function *ast.FuncType) bool {
	if len(function.Params.List) != 2 {
		return false
	}
	if fieldName(function.Params.List[0]) != "context.Context" {
		return false
	}
	return true
}

func hasValidOutputs(function *ast.FuncType) bool {
}

func ParseModelStruct(astOjb *ast.Object) *ServiceModelDeclaration {
	return &ServiceModelDeclaration{Name: astOjb.Name}
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

	// Your input must be a struct
	// TODO: support type aliases so you can re-use common inputs/outputs
	_, ok = typeSpec.Type.(*ast.StructType)
	if !ok {
		return false
	}

	// We're enforcing the convention that the input to the method "CreateFoo" must be "CreateFooRequest"
	// and the output must be called "CreateFooResponse"
	return strings.HasSuffix(astObj.Name, "Request") || strings.HasSuffix(astObj.Name, "Response")
}

type DeclarationContext struct {
	SourcePath string
	Services   []*ServiceDeclaration
	Models     []*ServiceModelDeclaration
}

func (ctx *DeclarationContext) ServiceByName(name string) *ServiceDeclaration {
	for _, service := range ctx.Services {
		if service.Name == name {
			return service
		}
	}
	return nil
}

func (ctx *DeclarationContext) ModelByName(name string) *ServiceModelDeclaration {
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
	Name     string
	Request  *ServiceModelDeclaration
	Response *ServiceModelDeclaration
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
}

func (model ServiceModelDeclaration) MethodName() string {
	name := model.Name
	name = strings.TrimSuffix(name, "Request")
	name = strings.TrimSuffix(name, "Response")
	return name
}

func (model ServiceModelDeclaration) String() string {
	return model.Name
}


*/

package parser_test

import (
	"testing"

	"github.com/monadicstack/frodo/parser"
	"github.com/stretchr/testify/suite"
)

/*
 * OTHER TESTS TO WRITE:
 * Models in different packages
 */
type ParserSuite struct {
	suite.Suite
}

// A baseline parsing test that has one service, one function, and some very simple request/response model structs.
func (suite *ParserSuite) TestBasic() {
	ctx, err := parser.ParseFile("testdata/basic/service.go")
	suite.Require().NoError(err)
	suite.assertContext(ctx, expectedContext{NumServices: 1, NumModels: 2})

	service := ctx.ServiceByName("DudeService")
	suite.assertService(service, expectedService{
		Name:         "DudeService",
		NumFunctions: 1,
	})

	suite.assertFunction(service, "Bowl", expectedFunction{
		RequestType:   "BowlRequest",
		ResponseType:  "BowlResponse",
		Documentation: parser.DocumentationLines{},
		Gateway:       expectedGateway{Method: "POST", Path: "/DudeService.Bowl", Status: 200},
	})

	model := ctx.ModelByName("BowlRequest")
	suite.assertModel(model, expectedModel{Name: "BowlRequest", NumFields: 1})
	suite.assertField(model, "BowlerID", expectedField{TypeName: "string"})

	model = ctx.ModelByName("BowlResponse")
	suite.assertModel(model, expectedModel{Name: "BowlResponse", NumFields: 2})
	suite.assertField(model, "BowlerID", expectedField{TypeName: "string"})
	suite.assertField(model, "Pins", expectedField{TypeName: "int"})
}

// Ensure that all of the doc options have the correct effect on the parsed context.
func (suite *ParserSuite) TestDocOptions() {
	ctx, err := parser.ParseFile("testdata/docoptions/service.go")
	suite.Require().NoError(err)
	suite.assertContext(ctx, expectedContext{NumServices: 1, NumModels: 2})

	service := ctx.ServiceByName("LebowskiService")
	suite.assertService(service, expectedService{
		Name:         "LebowskiService",
		Version:      "999.12",
		PathPrefix:   "/big",
		NumFunctions: 8,
	})

	suite.assertFunction(service, "Dude", expectedFunction{
		Documentation: parser.DocumentationLines{
			"Dude abides.",
		},
		Gateway: expectedGateway{Method: "GET", Path: "/dude/:id", Status: 202},
	})
	suite.assertFunction(service, "Walter", expectedFunction{
		Documentation: parser.DocumentationLines{},
		Gateway:       expectedGateway{Method: "POST", Path: "/LebowskiService.Walter", Status: 200},
	})
	suite.assertFunction(service, "Donnie", expectedFunction{
		Documentation: parser.DocumentationLines{},
		Gateway:       expectedGateway{Method: "POST", Path: "/LebowskiService.Donnie", Status: 204},
	})
	suite.assertFunction(service, "Maude", expectedFunction{
		Documentation: parser.DocumentationLines{},
		Gateway:       expectedGateway{Method: "POST", Path: "/dude/:id/child", Status: 201},
	})
	suite.assertFunction(service, "Jackie", expectedFunction{
		Documentation: parser.DocumentationLines{},
		Gateway:       expectedGateway{Method: "PUT", Path: "/dude/jail", Status: 200},
	})
	suite.assertFunction(service, "Stranger", expectedFunction{
		Documentation: parser.DocumentationLines{
			"Sometimes you eat the bar.",
			"",
			"Sometimes the bar eats you.",
		},
		Gateway: expectedGateway{Method: "PATCH", Path: "/dude/:id", Status: 200},
	})
	suite.assertFunction(service, "RemoveToe", expectedFunction{
		Documentation: parser.DocumentationLines{
			"RemoveToe attempts to extort $1 million.",
		},
		Gateway: expectedGateway{Method: "DELETE", Path: "/nihilist/:id/toe", Status: 200},
	})
	suite.assertFunction(service, "Rug", expectedFunction{
		Documentation: parser.DocumentationLines{
			"* HTTP 202",
		},
		Gateway: expectedGateway{Method: "HEAD", Path: "/ties/room/together", Status: 200},
	})
}

func (suite *ParserSuite) TestFieldTypes() {
	ctx, err := parser.ParseFile("testdata/fieldtypes/service.go")
	suite.Require().NoError(err)
	suite.assertContext(ctx, expectedContext{NumServices: 1, NumModels: 3})

	fields := ctx.ModelByName("Request").Fields

	suite.assertFieldType(fields, "Basic", expectedFieldType{Name: "string", Pointer: false, JSON: "string"})
	suite.assertFieldType(fields, "BasicPointer", expectedFieldType{Name: "*string", Pointer: true, JSON: "string"})

	suite.assertFieldType(fields, "ExportedStruct", expectedFieldType{Name: "ExportedStruct", Pointer: false, JSON: "object"})
	suite.assertFieldType(fields, "ExportedStructPointer", expectedFieldType{Name: "ExportedStruct", Pointer: true, JSON: "object"})

	suite.assertFieldType(fields, "NotExportedStruct", expectedFieldType{Name: "notExportedStruct", Pointer: false, JSON: "object"})
	suite.assertFieldType(fields, "NotExportedStructPointer", expectedFieldType{Name: "notExportedStruct", Pointer: true, JSON: "object"})

	suite.assertFieldType(fields, "Time", expectedFieldType{Name: "time.Time", Pointer: false, JSON: "string"})
	suite.assertFieldType(fields, "TimePointer", expectedFieldType{Name: "*time.Time", Pointer: true, JSON: "string"})

	suite.assertFieldType(fields, "Duration", expectedFieldType{Name: "time.Duration", Pointer: false, JSON: "number"})
	suite.assertFieldType(fields, "DurationPointer", expectedFieldType{Name: "*time.Duration", Pointer: true, JSON: "number"})

	suite.assertFieldType(fields, "BasicSlice", expectedFieldType{Name: "[]string", Pointer: false, JSON: "array", ElemName: "string"})
	suite.assertFieldType(fields, "BasicMap", expectedFieldType{Name: "map[string]string", Pointer: false, JSON: "array", ElemName: "string", KeyName: "string"})

	// TODO: Do rest of types...
}

// Ensures that you can have multiple services defined in the same file (though not preferred).
func (suite *ParserSuite) TestMultiService() {
	ctx, err := parser.ParseFile("testdata/multiservice/service.go")
	suite.Require().NoError(err)
	suite.assertContext(ctx, expectedContext{NumServices: 2, NumModels: 4})

	suite.assertService(ctx.ServiceByName("FooService"), expectedService{Name: "FooService", NumFunctions: 1})
	suite.assertService(ctx.ServiceByName("BarService"), expectedService{Name: "BarService", NumFunctions: 2})
}

func (suite *ParserSuite) TestErrorBadGoMod() {
	_, err := parser.ParseFile("testdata/errors/badgomod/service.go")
	suite.Require().Error(err, "Should fail when we can't properly process go.mod.")
	suite.Require().Contains(err.Error(), "go.mod")

	_, err = parser.ParseFile("testdata/errors/badgomod/foo/service.go")
	suite.Require().Error(err, "Should fail when we can't properly process go.mod recursively.")
	suite.Require().Contains(err.Error(), "go.mod")
}

func (suite *ParserSuite) TestErrorNoFile() {
	_, err := parser.ParseFile("testdata/does_not_exist.go")
	suite.Require().Error(err, "Should fail when file does not exist.")
}

func (suite *ParserSuite) TestErrorBlank() {
	_, err := parser.ParseFile("testdata/errors/blank/service.go")
	suite.Require().Error(err, "Should fail when file is blank")
}

func (suite *ParserSuite) TestErrorPackageOnly() {
	_, err := parser.ParseFile("testdata/errors/pkgonly/service.go")
	suite.Require().Error(err, "Should fail when file has nothing but the package identifier")
}

func (suite *ParserSuite) TestErrorNoServices() {
	_, err := parser.ParseFile("testdata/errors/noservices/service.go")
	suite.Require().Error(err, "Should fail when file does not contain any service interfaces")
}

// Ensure that we validate the service function parameters properly and that the error messages have
// something meaningful indicating why the function is invalid.
func (suite *ParserSuite) TestErrorFunctionParams() {
	_, err := parser.ParseFile("testdata/errors/paramcount/service.go")
	suite.Require().Error(err, "Should fail when a function does not have 2 params")
	suite.Require().Contains(err.Error(), "2", "Error should include the required parameter count")

	_, err = parser.ParseFile("testdata/errors/contextparam/service.go")
	suite.Require().Error(err, "Should fail when a function's first arg is not a Context")
	suite.Require().Contains(err.Error(), "context.Context", "Error should mention the context.Context")

	_, err = parser.ParseFile("testdata/errors/reqnotpointer/service.go")
	suite.Require().Error(err, "Should fail when a function's second arg is not a pointer")
	suite.Require().Contains(err.Error(), "pointer", "Error should mention the need for a pointer")

	_, err = parser.ParseFile("testdata/errors/reqnotstruct/service.go")
	suite.Require().Error(err, "Should fail when a function's second arg is not a struct in the file")
	suite.Require().Contains(err.Error(), "struct", "Error should mention the struct requirement")
}

func (suite *ParserSuite) TestErrorResponseNotStruct() {
	_, err := parser.ParseFile("testdata/errors/resultcount/service.go")
	suite.Require().Error(err, "Should fail when a function does not have 2 return values")
	suite.Require().Contains(err.Error(), "2", "Error should include the required return value count")

	_, err = parser.ParseFile("testdata/errors/resnotpointer/service.go")
	suite.Require().Error(err, "Should fail when a function's first return value is not a pointer")
	suite.Require().Contains(err.Error(), "pointer", "Error should mention the need for a pointer")

	_, err = parser.ParseFile("testdata/errors/resnotstruct/service.go")
	suite.Require().Error(err, "Should fail when a function's first return value is not a struct in the file")
	suite.Require().Contains(err.Error(), "struct", "Error should mention the struct requirement")

	_, err = parser.ParseFile("testdata/errors/resulterror/service.go")
	suite.Require().Error(err, "Should fail when a function's second return value is not an error")
	suite.Require().Contains(err.Error(), "error", "Error should mention the need for an error return value")
}

/*
 * ----------- Assertion Helpers ----------------------
 */

// assertContext checks the top-level parsing information; basically just making sure that we've captured the
// correct number of services and models.
func (suite *ParserSuite) assertContext(ctx *parser.Context, expected expectedContext) {
	suite.Require().NotNil(ctx, "Context should not be nil")
	suite.Require().Equal(expected.NumServices, len(ctx.Services), "Incorrect service count")
	suite.Require().Equal(expected.NumModels, len(ctx.Models), "Incorrect model count")
}

// assertService makes sure that a service declaration from the parsed context contains all of the
// attributes we would expect given the input source code. It does not perform any assertions on the
// functions of the service other than the count; those need to be done separately.
func (suite *ParserSuite) assertService(s *parser.ServiceDeclaration, expected expectedService) {
	if expected.Version == "" {
		expected.Version = parser.DefaultServiceVersion
	}
	name := expected.Name
	suite.Require().NotNil(s, "%s: Not found")
	suite.Require().Equal(expected.Name, s.Name, "%s: Incorrect name", name)
	suite.Require().Equal(expected.NumFunctions, len(s.Functions), "%s: Incorrect function count", name)
	suite.Require().Equal(expected.Version, s.Version, "%s: Incorrect version", name)
	suite.Require().NotNil(s.Gateway, "%s: Missing gateway options", name)
	suite.Require().Equal(expected.PathPrefix, s.Gateway.PathPrefix, "%s: Incorrect version", name)
}

// assertFunction makes sure that a service function declaration contains all of the attributes
// we would expect given the input source code. This includes checking customizations from Doc Options.
func (suite *ParserSuite) assertFunction(service *parser.ServiceDeclaration, functionName string, expected expectedFunction) {
	f := service.FunctionByName(functionName)
	name := service.Name + "." + functionName + "()"

	suite.Require().NotNil(f, "%s: Not found", name)
	suite.Require().Equal(functionName, f.Name, "%s: Incorrect name", name)
	suite.Require().Equal(service, f.Service, "%s: Incorrect service back-pointer", name)
	suite.Require().Equal(expected.Documentation.String(), f.Documentation.String(), "%s: Incorrect documentation", name)

	gateway := f.Gateway
	suite.Require().NotNil(gateway, "%s: Gateway: Not found", name)
	suite.Require().Equal(f, gateway.Function, "%s: Gateway: Incorrect function back-pointer", name)
	suite.Require().Equal(expected.Gateway.Path, gateway.Path, "%s: Gateway: Incorrect path", name)
	suite.Require().Equal(expected.Gateway.Method, gateway.Method, "%s: Gateway: Incorrect method", name)
	suite.Require().Equal(expected.Gateway.Status, gateway.Status, "%s: Gateway: Incorrect status", name)

	// Only check the model types if specified. Blank means this test doesn't care about the request/response models.
	if expected.RequestType != "" {
		request := f.Request
		suite.Require().NotNil(request, "%s: Request Struct: Not found", name)
		suite.Require().Equal(expected.RequestType, request.Type.Name, "%s: Request Struct: Wrong type name", name)
	}
	if expected.ResponseType != "" {
		response := f.Response
		suite.Require().NotNil(response, "%s: Response Struct: Not found", name)
		suite.Require().Equal(expected.ResponseType, response.Type.Name, "%s: Response Struct: Wrong type name", name)
	}
}

func (suite *ParserSuite) assertModel(m *parser.ServiceModelDeclaration, expected expectedModel) {
	name := expected.Name
	suite.Require().NotNil(m, "%s: Not found", name)
	suite.Require().Equal(expected.Name, m.Name, "%s: Incorrect name", name)
	suite.Require().Equal(expected.Documentation.String(), m.Documentation.String(), "%s: Incorrect documentation", name)
	suite.Require().Equal(expected.NumFields, len(m.Fields), "%s: Incorrect field count", name)

	fieldType := m.Type
	suite.Require().NotNil(fieldType, "%s: Type: Not found", name)
	suite.Require().False(fieldType.Pointer, "%s: Type: Should not be a pointer", name)
}

func (suite *ParserSuite) assertField(model *parser.ServiceModelDeclaration, fieldName string, expected expectedField) {
	f := model.Fields.ByName(fieldName)
	name := model.Name + "." + fieldName

	suite.Require().NotNil(f, "%s: Not found", name)
	suite.Require().Equal(fieldName, f.Name, "%s: Incorrect name", name)
	suite.Require().Equal(model, f.Model, "%s: Incorrect model back-pointer", name)
	suite.Require().Equal(expected.Documentation.String(), f.Documentation.String(), "%s: Incorrect documentation", name)

	fieldType := f.Type
	suite.Require().NotNil(fieldType, "%s: Type: Not found", name)
	suite.Require().Equal(expected.TypeName, fieldType.Name, "%s: Type: Incorrect name", name)
	suite.Require().Equal(expected.TypePointer, fieldType.Pointer, "%s: Type: Incorrect pointer flag", name)
}

func (suite *ParserSuite) assertFieldType(fields parser.FieldDeclarations, name string, expected expectedFieldType) {
	field := fields.ByName(name)
	suite.Require().NotNil(field, "%s: Field is missing altogether", name)

	t := field.Type
	suite.Require().Equal(expected.Name, t.Name, "%s: Incorrect name", name)
	suite.Require().Equal(expected.Pointer, t.Pointer, "%s: Incorrect pointer flag", name)

	if expected.ElemName == "" {
		suite.Require().Nil(t.Elem, "%s: Elem type should be nil", name)
	} else {
		suite.Require().NotNil(t.Elem, "%s: Elem type not found", name)
		suite.Require().Equal(expected.ElemName, t.Elem.Name, "%s: Incorrect 'elem' type name", name)
	}

	if expected.KeyName == "" {
		suite.Require().Nil(t.Key, "%s: Key type should be nil", name)
	} else {
		suite.Require().NotNil(t.Key, "%s: Key type not found", name)
		suite.Require().Equal(expected.KeyName, t.Key.Name, "%s: Incorrect 'key' type name", name)
	}
}

type expectedContext struct {
	NumServices int
	NumModels   int
}

type expectedService struct {
	Name         string
	NumFunctions int
	Version      string
	PathPrefix   string
}

type expectedFunction struct {
	RequestType   string
	ResponseType  string
	Documentation parser.DocumentationLines
	Gateway       expectedGateway
}

type expectedGateway struct {
	Path   string
	Method string
	Status int
}

type expectedModel struct {
	Name          string
	NumFields     int
	Documentation parser.DocumentationLines
}

type expectedField struct {
	TypeName      string
	TypePointer   bool
	Documentation parser.DocumentationLines
}

type expectedFieldType struct {
	Name     string
	Pointer  bool
	ElemName string
	KeyName  string
	JSON     string
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

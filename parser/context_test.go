// +build unit

package parser_test

import (
	"go/ast"
	"strings"
	"testing"
	"time"

	"github.com/monadicstack/frodo/parser"
	"github.com/stretchr/testify/suite"
)

type ContextSuite struct {
	suite.Suite
}

func (suite *ContextSuite) TestService_FunctionByName() {
	service := &parser.ServiceDeclaration{}
	check := func(service *parser.ServiceDeclaration, name string, exists bool) {
		function := service.FunctionByName(name)
		if exists {
			suite.Require().NotNil(function, "Service did not find function '%s'", name)
			suite.Require().True(strings.EqualFold(name, function.Name), "Service found wrong function '%s'", name)
		} else {
			suite.Require().Nil(function, "Service should not find the function '%s'", name)
		}
	}

	check(service, "", false)
	check(service, "Foo", false)

	service.Functions = []*parser.ServiceFunctionDeclaration{
		{Name: "Foo"},
	}
	check(service, "", false)
	check(service, "Foo", true)
	check(service, "foo", true)
	check(service, "FooFunc", false)
	check(service, "Bar", false)

	service.Functions = []*parser.ServiceFunctionDeclaration{
		{Name: "Foo"},
		{Name: "Bar"},
	}
	check(service, "", false)
	check(service, "Foo", true)
	check(service, "FooFunc", false)
	check(service, "Bar", true)
	check(service, "bar", true)
	check(service, "Baz", false)
}

func (suite *ContextSuite) TestFunction_String() {
	function := &parser.ServiceFunctionDeclaration{}
	suite.Require().Equal("(context.Context, *<nil>) (*<nil>, error)", function.String())

	function = &parser.ServiceFunctionDeclaration{
		Name:     "Foo",
		Request:  &parser.TypeDeclaration{Name: "Request"},
		Response: &parser.TypeDeclaration{Name: "Response"},
	}
	suite.Require().Equal("Foo(context.Context, *Request) (*Response, error)", function.String())
}

func (suite *ContextSuite) TestFieldDeclarations_Empty_NotEmpty() {
	fields := parser.FieldDeclarations{}
	suite.Require().True(fields.Empty())
	suite.Require().False(fields.NotEmpty())

	fields = parser.FieldDeclarations{{}}
	suite.Require().False(fields.Empty())
	suite.Require().True(fields.NotEmpty())

	fields = parser.FieldDeclarations{{Name: "Foo"}, {Name: "Bar"}}
	suite.Require().False(fields.Empty())
	suite.Require().True(fields.NotEmpty())
}

func (suite *ContextSuite) TestFieldDeclarations_ByName() {
	check := func(fields parser.FieldDeclarations, name string, exists bool) {
		field := fields.ByName(name)
		if exists {
			suite.Require().NotNil(field, "Did not find field '%s'", name)
			suite.Require().True(strings.EqualFold(name, field.Name), "Found wrong field '%s'", name)
		} else {
			suite.Require().Nil(field, "Should not find the field '%s'", name)
		}
	}

	fields := parser.FieldDeclarations{}
	check(fields, "", false)
	check(fields, "Foo", false)

	fields = parser.FieldDeclarations{{Name: "Foo"}}
	check(fields, "", false)
	check(fields, "Foo", true)
	check(fields, "FooFunc", false)
	check(fields, "foo", true)
	check(fields, "Bar", false)

	fields = parser.FieldDeclarations{{Name: "Foo"}, {Name: "Bar"}}
	check(fields, "", false)
	check(fields, "Foo", true)
	check(fields, "FooFunc", false)
	check(fields, "foo", true)
	check(fields, "Bar", true)
	check(fields, "bar", true)
	check(fields, "bbar", false)
}

func (suite *ContextSuite) TestFieldDeclarations_ByBindingName() {
	newField := func(name string, bindingName string) *parser.FieldDeclaration {
		return &parser.FieldDeclaration{
			Name:    name,
			Binding: &parser.FieldBindingOptions{Name: bindingName},
		}
	}
	check := func(fields parser.FieldDeclarations, name string, exists bool) {
		field := fields.ByBindingName(name)
		if exists {
			suite.Require().NotNil(field, "Did not find field '%s'", name)
			suite.Require().True(strings.EqualFold(name, field.Binding.Name), "Found wrong field '%s'", name)
		} else {
			suite.Require().Nil(field, "Should not find the field '%s'", name)
		}
	}

	fields := parser.FieldDeclarations{}
	check(fields, "", false)
	check(fields, "Foo", false)

	fields = parser.FieldDeclarations{newField("Foo", "Foo")}
	check(fields, "", false)
	check(fields, "Foo", true)
	check(fields, "FooFunc", false)
	check(fields, "foo", true)
	check(fields, "Bar", false)

	fields = parser.FieldDeclarations{newField("Foo", "Goo")}
	check(fields, "", false)
	check(fields, "Foo", false)
	check(fields, "Goo", true)
	check(fields, "goo", true)
	check(fields, "Bar", false)

	fields = parser.FieldDeclarations{newField("Foo", "Goo"), newField("Bar", "Baz")}
	check(fields, "", false)
	check(fields, "Foo", false)
	check(fields, "Goo", true)
	check(fields, "goo", true)
	check(fields, "Bar", false)
	check(fields, "Baz", true)
	check(fields, "baz", true)
}

func (suite *ContextSuite) TestBindingOptions_NotOmit() {
	binding := parser.FieldBindingOptions{}
	suite.Require().True(binding.NotOmit())

	binding.Omit = false
	suite.Require().True(binding.NotOmit())

	binding.Omit = true
	suite.Require().False(binding.NotOmit())
}

func (suite *ContextSuite) TestModuleDeclaration_GoMod() {
	module := parser.ModuleDeclaration{}
	suite.Require().Equal("go.mod", module.GoMod())

	module.Directory = "."
	suite.Require().Equal("go.mod", module.GoMod())

	module.Directory = "foo"
	suite.Require().Equal("foo/go.mod", module.GoMod())

	module.Directory = "../foo/bar"
	suite.Require().Equal("../foo/bar/go.mod", module.GoMod())

	module.Directory = "/absolutely/this/works"
	suite.Require().Equal("/absolutely/this/works/go.mod", module.GoMod())
}

func (suite *ContextSuite) TestDocumentation_Set() {
	// Need at least two arguments for anything to be added to the map
	docs := parser.Documentation{}
	docs.Set()
	docs.Set("a")
	suite.Require().Len(docs, 0)

	// We currently don't do any special cleanup if you provide empty segments.
	docs.Set("", "", "", "junk")
	suite.Require().Equal("junk", docs[".."])

	docs.Set("a", "b", "Hello")
	suite.Require().Equal("Hello", docs["a.b"])

	docs.Set("a", "b", "c", "World")
	suite.Require().Equal("", docs["a"])
	suite.Require().Equal("Hello", docs["a.b"])
	suite.Require().Equal("World", docs["a.b.c"])

	docs.Set("a", "b", "c.d", "Goodbye")
	suite.Require().Equal("", docs["a"])
	suite.Require().Equal("Hello", docs["a.b"])
	suite.Require().Equal("World", docs["a.b.c"])
	suite.Require().Equal("Goodbye", docs["a.b.c.d"])
}

// Service and model doc lookups have the exact same semantics, so we make sure that if a docs map works for
// one it works for the other as long as the names line up.
func (suite *ContextSuite) TestDocumentation_ForService_ForModel() {
	checkService := func(docs parser.Documentation, name string, expectedLines ...string) {
		service := &parser.ServiceDeclaration{Name: name}
		serviceDocs := docs.ForService(service)
		suite.Require().Equal(parser.DocumentationLines(expectedLines).String(), serviceDocs.String())
	}
	checkModel := func(docs parser.Documentation, name string, expectedLines ...string) {
		model := &parser.TypeDeclaration{Name: name}
		modelDocs := docs.ForType(model)
		suite.Require().Equal(parser.DocumentationLines(expectedLines).String(), modelDocs.String())
	}
	check := func(docs parser.Documentation, name string, expectedLines ...string) {
		checkService(docs, name, expectedLines...)
		checkModel(docs, name, expectedLines...)
	}

	docs := parser.Documentation{}

	check(docs, "ServiceA", "")

	docs.Set("ServiceA", "Comment A")

	check(docs, "ServiceA", "Comment A")
	check(docs, "ServiceB")

	docs.Set("ServiceA", "Func1", "Comment A1")
	docs.Set("ServiceA", "Func2", "Comment A2")
	docs.Set("ServiceB", "Comment B\nComment B Line 2")

	check(docs, "ServiceA", "Comment A")
	check(docs, "ServiceB", "Comment B", "Comment B Line 2")
}

// Function and field doc lookups have the exact same semantics, so we make sure that if a docs map works for
// one it works for the other as long as the names line up.
func (suite *ContextSuite) TestDocumentation_ForFunction_ForField() {
	checkFunction := func(docs parser.Documentation, serviceName string, name string, expectedLines ...string) {
		service := &parser.ServiceDeclaration{Name: serviceName}
		function := &parser.ServiceFunctionDeclaration{Name: name, Service: service}
		functionDocs := docs.ForFunction(function)
		suite.Require().Equal(parser.DocumentationLines(expectedLines).String(), functionDocs.String())
	}
	checkField := func(docs parser.Documentation, modelName string, name string, expectedLines ...string) {
		model := &parser.TypeDeclaration{Name: modelName}
		field := &parser.FieldDeclaration{Name: name, ParentType: model}
		fieldDocs := docs.ForField(field)
		suite.Require().Equal(parser.DocumentationLines(expectedLines).String(), fieldDocs.String())
	}
	check := func(docs parser.Documentation, parentName string, name string, expectedLines ...string) {
		checkFunction(docs, parentName, name, expectedLines...)
		checkField(docs, parentName, name, expectedLines...)
	}

	docs := parser.Documentation{}

	check(docs, "ServiceA.Func1", "")

	docs.Set("ServiceA", "Func1", "Comment A1")
	docs.Set("ServiceA", "Func2", "Comment A2")
	docs.Set("ServiceB", "Comment B\nComment B Line 2")
	docs.Set("ServiceA", "Func1", "Comment B1")

	check(docs, "ServiceA.Func1", "Comment A1")
	check(docs, "ServiceA.Func2", "Comment A2")
	check(docs, "ServiceB.Func1", "Comment B1")
	check(docs, "ServiceB.Func1.Other", "")
}

func (suite *ContextSuite) TestTags_Set() {
	// Need at least two arguments for anything to be added to the map
	docs := parser.Tags{}
	docs.Set("", "", &ast.BasicLit{Value: "0"})
	docs.Set("x", "", &ast.BasicLit{Value: "X"})
	docs.Set("", "y", &ast.BasicLit{Value: "Y"})
	suite.Require().Len(docs, 0)

	// We trim off the back-ticks.
	docs.Set("a", "b", &ast.BasicLit{Value: "`json:\"foo\"`"})
	suite.Require().Len(docs, 1)
	suite.Require().Equal("json:\"foo\"", docs["a.b"])
	suite.Require().Equal("", docs["a.b.c"])

	// Replacing a value actually replaces, not appends
	docs.Set("a", "b", &ast.BasicLit{Value: "`json:\"bar\"`"})
	suite.Require().Len(docs, 1)
	suite.Require().Equal("json:\"bar\"", docs["a.b"])
	suite.Require().Equal("", docs["a.b.c"])

	// Replacing a value actually replaces, not appends
	docs.Set("x", "y", &ast.BasicLit{Value: "`json:\"baz\"`"})
	suite.Require().Len(docs, 2)
	suite.Require().Equal("json:\"baz\"", docs["x.y"])
}

func (suite *ContextSuite) TestTags_ForField() {
	check := func(tags parser.Tags, modelName string, fieldName string, expectedTag string) {
		model := &parser.TypeDeclaration{Name: modelName}
		field := &parser.FieldDeclaration{Name: fieldName, ParentType: model}
		tag := tags.ForField(field)
		suite.Require().Equal(expectedTag, string(tag))
	}

	tags := parser.Tags{}
	check(tags, "a", "b", "")

	tags.Set("a", "b", &ast.BasicLit{Value: "`json:\"foo\"`"})
	check(tags, "a", "b", "json:\"foo\"")
	check(tags, "", "a.b", "")

	tags.Set("a", "c", &ast.BasicLit{Value: "`json:\"bar\"`"})
	check(tags, "a", "b", "json:\"foo\"")
	check(tags, "a", "c", "json:\"bar\"")

	tags.Set("x", "y", &ast.BasicLit{Value: "`json:\"baz\"`"})
	check(tags, "a", "b", "json:\"foo\"")
	check(tags, "a", "c", "json:\"bar\"")
	check(tags, "x", "y", "json:\"baz\"")
}

func (suite *ContextSuite) TestDocumentationLines_Empty_NotEmpty() {
	docs := parser.DocumentationLines{}
	suite.Require().True(docs.Empty())
	suite.Require().False(docs.NotEmpty())

	docs = parser.DocumentationLines{""}
	suite.Require().True(docs.Empty())
	suite.Require().False(docs.NotEmpty())

	docs = parser.DocumentationLines{"foo"}
	suite.Require().False(docs.Empty())
	suite.Require().True(docs.NotEmpty())

	docs = parser.DocumentationLines{"", "foo", ""}
	suite.Require().False(docs.Empty())
	suite.Require().True(docs.NotEmpty())

	docs = parser.DocumentationLines{"", "foo", "", "baz"}
	suite.Require().False(docs.Empty())
	suite.Require().True(docs.NotEmpty())
}

func (suite *ContextSuite) TestDocumentationLines_Trim() {
	docs := parser.DocumentationLines{}.Trim()
	suite.Require().Equal("", docs.String())

	docs = parser.DocumentationLines{""}.Trim()
	suite.Require().Equal("", docs.String())

	docs = parser.DocumentationLines{"foo"}.Trim()
	suite.Require().Equal("foo", docs.String())

	docs = parser.DocumentationLines{"", "foo", ""}.Trim()
	suite.Require().Equal("foo", docs.String())

	docs = parser.DocumentationLines{"", "foo", "", "baz", "", "   \n", ""}.Trim()
	suite.Require().Equal("foo\n\nbaz", docs.String())
}

func (suite *ContextSuite) TestGatewayFunctionOptions_SupportsBody() {
	check := func(method string, expected bool) {
		options := parser.GatewayFunctionOptions{Method: method}
		suite.Require().Equal(expected, options.SupportsBody(), "Gateway.SupportsBody(%s) should be %v", method, expected)
	}

	check("", false)
	check("fart", false)
	check("GET", false)
	check("get", false)
	check("DELETE", false)
	check("HEAD", false)
	check("OPTIONS", false)

	check("POST", true)
	check("post", true)
	check("Post", true)
	check("PUT", true)
	check("put", true)
	check("Put", true)
	check("PATCH", true)
	check("patch", true)
	check("Patch", true)
}

func (suite *ContextSuite) TestGatewayFunctionOptions_PathParameters() {
	fields := parser.FieldDeclarations{
		&parser.FieldDeclaration{Name: "ID", Binding: &parser.FieldBindingOptions{Name: "ID"}},
		&parser.FieldDeclaration{Name: "LastName", Binding: &parser.FieldBindingOptions{Name: "LastName"}},
		&parser.FieldDeclaration{Name: "FirstName", Binding: &parser.FieldBindingOptions{Name: "first_name"}},
	}
	request := &parser.TypeDeclaration{
		Fields: fields,
	}
	function := &parser.ServiceFunctionDeclaration{
		Request: request,
	}
	options := parser.GatewayFunctionOptions{
		Function: function,
	}

	options.Path = "/"
	params := options.PathParameters()
	suite.Require().Len(params, 0)

	options.Path = "/SomeService.SomeFunction"
	params = options.PathParameters()
	suite.Require().Len(params, 0)

	options.Path = "/foo/bar/baz"
	params = options.PathParameters()
	suite.Require().Len(params, 0)

	options.Path = ":id"
	params = options.PathParameters()
	suite.Require().Len(params, 1)
	suite.Require().Equal("id", params[0].Name)
	suite.Require().Equal("ID", params[0].Field.Name)

	// Case insensitive matches
	options.Path = "/foo/:id/baz/:FIRST_name"
	params = options.PathParameters()
	suite.Require().Len(params, 2)
	suite.Require().Equal("id", params[0].Name)
	suite.Require().Equal("ID", params[0].Field.Name)
	suite.Require().Equal("FIRST_name", params[1].Name)
	suite.Require().Equal("FirstName", params[1].Field.Name)

	// We don't include path params if there's no field on that request model
	options.Path = "/foo/:bar/baz"
	params = options.PathParameters()
	suite.Require().Len(params, 0)

	options.Path = "/foo/:id/bar/:first_name/baz/:middle_name"
	params = options.PathParameters()
	suite.Require().Len(params, 2)
	suite.Require().Equal("id", params[0].Name)
	suite.Require().Equal("ID", params[0].Field.Name)
	suite.Require().Equal("first_name", params[1].Name)
	suite.Require().Equal("FirstName", params[1].Field.Name)
}

func (suite *ContextSuite) TestGatewayFunctionOptions_QueryParameters() {
	fields := parser.FieldDeclarations{
		&parser.FieldDeclaration{Name: "ID", Binding: &parser.FieldBindingOptions{Name: "ID"}},
		&parser.FieldDeclaration{Name: "LastName", Binding: &parser.FieldBindingOptions{Name: "LastName"}},
		&parser.FieldDeclaration{Name: "FirstName", Binding: &parser.FieldBindingOptions{Name: "first_name"}},
	}
	request := &parser.TypeDeclaration{
		Fields: fields,
	}
	function := &parser.ServiceFunctionDeclaration{
		Request: request,
	}
	options := parser.GatewayFunctionOptions{
		Function: function,
	}

	checkParam := func(params parser.GatewayParameters, which int, expectedName string, expectedFieldName string) {
		suite.Require().Equal(expectedName, params[which].Name)
		suite.Require().Equal(expectedFieldName, params[which].Field.Name)
	}

	options.Path = "/"
	params := options.QueryParameters()
	suite.Require().Len(params, 3)
	checkParam(params, 0, "ID", "ID")
	checkParam(params, 1, "LastName", "LastName")
	checkParam(params, 2, "first_name", "FirstName")

	options.Path = "/SomeService.SomeFunction"
	params = options.QueryParameters()
	suite.Require().Len(params, 3)
	checkParam(params, 0, "ID", "ID")
	checkParam(params, 1, "LastName", "LastName")
	checkParam(params, 2, "first_name", "FirstName")

	// Having unrelated path params shouldn't take away from the available fields.
	options.Path = "/foo/:bar/baz"
	params = options.QueryParameters()
	suite.Require().Len(params, 3)
	checkParam(params, 0, "ID", "ID")
	checkParam(params, 1, "LastName", "LastName")
	checkParam(params, 2, "first_name", "FirstName")

	// If it's in the path, it should not be a query param.
	options.Path = "/foo/:id/baz"
	params = options.QueryParameters()
	suite.Require().Len(params, 2)
	checkParam(params, 0, "LastName", "LastName")
	checkParam(params, 1, "first_name", "FirstName")

	options.Path = "/foo/:id/baz/:first_name"
	params = options.QueryParameters()
	suite.Require().Len(params, 1)
	checkParam(params, 0, "LastName", "LastName")

	options.Path = "/foo/:id/baz/:first_name/:LastName"
	params = options.QueryParameters()
	suite.Require().Len(params, 0)
}

func (suite *ContextSuite) TestGatewayParameters_Empty_NotEmpty() {
	params := parser.GatewayParameters{}
	suite.Require().True(params.Empty())
	suite.Require().False(params.NotEmpty())

	params = append(params, &parser.GatewayParameter{})
	suite.Require().False(params.Empty())
	suite.Require().True(params.NotEmpty())

	params = append(params, &parser.GatewayParameter{Name: "Fart"})
	suite.Require().False(params.Empty())
	suite.Require().True(params.NotEmpty())
}

func (suite *ContextSuite) TestGatewayParameters_ByName() {
	params := parser.GatewayParameters{}
	check := func(params parser.GatewayParameters, name string, exists bool) {
		param := params.ByName(name)
		if exists {
			suite.Require().NotNil(param, "Did not find param '%s'", name)
			suite.Require().True(strings.EqualFold(name, param.Name), "Found wrong param '%s'", name)
		} else {
			suite.Require().Nil(param, "Should not find the param '%s'", name)
		}
	}

	check(params, "", false)
	check(params, "Foo", false)

	params = append(params, &parser.GatewayParameter{Name: "Foo"})
	check(params, "", false)
	check(params, "Foo", true)
	check(params, "foo", true)
	check(params, "Food", false)
	check(params, "Bar", false)

	params = append(params, &parser.GatewayParameter{Name: "Bar"})
	check(params, "", false)
	check(params, "Foo", true)
	check(params, "foo", true)
	check(params, "Food", false)
	check(params, "Bar", true)
}

func (suite *ContextSuite) TestContext_timestampString() {
	ctx := parser.Context{}
	suite.Equal("Mon, 01 Jan 0001 00:00:00 UTC", ctx.TimestampString())

	// We only worry about second-level granularity.
	ctx.Timestamp = time.Date(2021, time.September, 24, 14, 44, 42, 12345, time.UTC)
	suite.Equal("Fri, 24 Sep 2021 14:44:42 UTC", ctx.TimestampString())

	// We abide the location / time zone
	loc, _ := time.LoadLocation("America/New_York")
	ctx.Timestamp = time.Date(2021, time.September, 24, 14, 44, 42, 12345, loc)
	suite.Equal("Fri, 24 Sep 2021 14:44:42 EDT", ctx.TimestampString())
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextSuite))
}

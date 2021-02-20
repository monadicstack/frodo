package metadata_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/monadicstack/frodo/rpc/metadata"
	"github.com/stretchr/testify/suite"
)

type ValuesSuite struct {
	suite.Suite
}

// Ensures that .Value() and .WithValue() behave properly when interacting with nil/empty contexts.
func (suite *ValuesSuite) TestValues_emptyContext() {
	ctx := context.Background()

	// Should fail w/ false rather than panic w/ nil context
	suite.assertString(nil, testCase{key: "string", expect: "", expectOK: false})
	suite.assertInt(nil, testCase{key: "int", expect: 0, expectOK: false})
	suite.assertBool(nil, testCase{key: "bool", expect: false, expectOK: false})

	// Empty keys shouldn't panic the operation either.
	suite.assertString(ctx, testCase{key: "", expect: "", expectOK: false})
	suite.assertInt(ctx, testCase{key: "", expect: 0, expectOK: false})
	suite.assertBool(ctx, testCase{key: "", expect: false, expectOK: false})

	// Make sure non-existent keys leave variables at their defaults
	suite.assertString(ctx, testCase{key: "string", expect: "", expectOK: false})
	suite.assertInt(ctx, testCase{key: "int", expect: 0, expectOK: false})
	suite.assertBool(ctx, testCase{key: "bool", expect: false, expectOK: false})

	// Inserting values with junk data doesn't panic either
	suite.Nil(metadata.WithValue(nil, "string", "12345"))
	suite.Equal(ctx, metadata.WithValue(ctx, "", "12345"))
}

func (suite *ValuesSuite) TestValues_valueTypes() {
	ctx := context.Background()
	ctx = metadata.WithValue(ctx, "string", "12345")
	ctx = metadata.WithValue(ctx, "int", 9999)
	ctx = metadata.WithValue(ctx, "bool", true)
	ctx = metadata.WithValue(ctx, "struct", structValue{Name: "Kid", Age: 12})

	// Non-existent keys should still give nothing back.
	suite.assertString(ctx, testCase{key: "", expect: "", expectOK: false})
	suite.assertString(ctx, testCase{key: "dragon", expect: "", expectOK: false})

	// After adding some values, make sure that we can look them up properly.
	suite.assertString(ctx, testCase{key: "string", expect: "12345", expectOK: true})
	suite.assertInt(ctx, testCase{key: "int", expect: 9999, expectOK: true})
	suite.assertBool(ctx, testCase{key: "bool", expect: true, expectOK: true})
	suite.assertStruct(ctx, testCase{key: "struct", expect: structValue{Name: "Kid", Age: 12}, expectOK: true})

	// Changing values on the context shouldn't be a problem
	ctx = metadata.WithValue(ctx, "string", "67890")
	ctx = metadata.WithValue(ctx, "int", 7777)
	ctx = metadata.WithValue(ctx, "struct", structValue{Name: "Dude", Age: 47})
	suite.assertString(ctx, testCase{key: "string", expect: "67890", expectOK: true})
	suite.assertInt(ctx, testCase{key: "int", expect: 7777, expectOK: true})
	suite.assertBool(ctx, testCase{key: "bool", expect: true, expectOK: true})
	suite.assertStruct(ctx, testCase{key: "struct", expect: structValue{Name: "Dude", Age: 47}, expectOK: true})
}

func (suite *ValuesSuite) TestValues_pointerTypes() {
	// Should work if we put pointers in and ask for pointers back
	ctx := context.Background()
	ctx = metadata.WithValue(ctx, "string", stringPtr("12345"))
	ctx = metadata.WithValue(ctx, "int", intPtr(9999))
	ctx = metadata.WithValue(ctx, "bool", boolPtr(true))
	ctx = metadata.WithValue(ctx, "struct", &structValue{Name: "Kid", Age: 12})
	suite.assertStringPtr(ctx, testCase{key: "string", expect: stringPtr("12345"), expectOK: true})
	suite.assertIntPtr(ctx, testCase{key: "int", expect: intPtr(9999), expectOK: true})
	suite.assertBoolPtr(ctx, testCase{key: "bool", expect: boolPtr(true), expectOK: true})
	suite.assertStructPtr(ctx, testCase{key: "struct", expect: &structValue{Name: "Kid", Age: 12}, expectOK: true})

	// Should work if we put values in and ask for pointers back
	ctx = context.Background()
	ctx = metadata.WithValue(ctx, "string", "12345")
	ctx = metadata.WithValue(ctx, "int", 9999)
	ctx = metadata.WithValue(ctx, "bool", true)
	ctx = metadata.WithValue(ctx, "struct", structValue{Name: "Kid", Age: 12})
	suite.assertStringPtr(ctx, testCase{key: "string", expect: stringPtr("12345"), expectOK: true})
	suite.assertIntPtr(ctx, testCase{key: "int", expect: intPtr(9999), expectOK: true})
	suite.assertBoolPtr(ctx, testCase{key: "bool", expect: boolPtr(true), expectOK: true})
	suite.assertStructPtr(ctx, testCase{key: "struct", expect: &structValue{Name: "Kid", Age: 12}, expectOK: true})
}

// Silently return ok=false when you provide an &out param that is the wrong type.
func (suite *ValuesSuite) TestValues_wrongType() {
	ctx := context.Background()
	ctx = metadata.WithValue(ctx, "string", "12345")
	ctx = metadata.WithValue(ctx, "bool", true)

	suite.assertString(ctx, testCase{key: "bool", expect: "", expectOK: false})
	suite.assertBool(ctx, testCase{key: "string", expect: false, expectOK: false})
}

// Ensure that we can marshal/unmarshal metadata values as JSON so they can be transported between contexts.
func (suite *ValuesSuite) TestValues_json() {
	a := context.Background()
	a = metadata.WithValue(a, "string", "12345")
	a = metadata.WithValue(a, "int", 9999)
	a = metadata.WithValue(a, "bool", true)
	a = metadata.WithValue(a, "struct", structValue{Name: "Kid", Age: 12})

	b := context.Background()
	suite.assertString(b, testCase{key: "string", expect: "", expectOK: false})
	suite.assertInt(b, testCase{key: "int", expect: 0, expectOK: false})
	suite.assertBool(b, testCase{key: "bool", expect: false, expectOK: false})
	suite.assertStruct(b, testCase{key: "struct", expect: structValue{}, expectOK: false})

	valueJSON, err := metadata.ToJSON(a)
	suite.Require().NoError(err)

	fmt.Println(valueJSON)
	values, err := metadata.FromJSON(valueJSON)
	suite.Require().NoError(err)

	b = metadata.WithValues(b, values)
	suite.assertString(b, testCase{key: "string", expect: "12345", expectOK: true})
	suite.assertInt(b, testCase{key: "int", expect: 9999, expectOK: true})
	suite.assertBool(b, testCase{key: "bool", expect: true, expectOK: true})
	suite.assertStruct(b, testCase{key: "struct", expect: structValue{Name: "Kid", Age: 12}, expectOK: true})
}

// When transporting metadata from a context w/ no values, don't fail, just keep it empty.
func (suite *ValuesSuite) TestValues_json_noValues() {
	a := context.Background()

	valueJSON, err := metadata.ToJSON(a)
	suite.Require().NoError(err)

	values, err := metadata.FromJSON(valueJSON)
	suite.Require().NoError(err)

	b := metadata.WithValues(context.Background(), values)
	suite.assertString(b, testCase{key: "string", expect: "", expectOK: false})
	suite.assertInt(b, testCase{key: "int", expect: 0, expectOK: false})
	suite.assertBool(b, testCase{key: "bool", expect: false, expectOK: false})
	suite.assertStruct(b, testCase{key: "struct", expect: structValue{}, expectOK: false})
}

func (suite *ValuesSuite) TestValues_json_invalid() {
	var ctx context.Context

	values, err := metadata.FromJSON("")
	suite.Require().NoError(err, "Should not return an error when FromJSON receives ''.")
	suite.Require().Len(values, 0, "Should return an empty values when FromJSON receives ''.")

	values, err = metadata.FromJSON(`{garbage`)
	suite.Require().Error(err, "Should return an error when FromJSON receives garbage.")
	suite.Require().Len(values, 0, "Should return an empty values when FromJSON receives garbage.")

	// Age should be an int, not a bool as the JSON would have you believe. This should not generate
	// an error, but we won't be able to properly fetch the "struct" value from the context later. Basically
	// the time we do the parsing, we don't even really know what the target type is. So we won't know
	// that "Age" is garbage until we call `metadata.Value()` where we *do* have type info.
	values, err = metadata.FromJSON(`{"struct":{"value":{"Name":"Kid","Age":true}}, "string":false}`)
	suite.NoError(err)
	suite.Require().Len(values, 2, "Should keep metadata entries even if JSON is not correct for type.")

	ctx = context.Background()
	ctx = metadata.WithValues(ctx, values)
	suite.assertString(ctx, testCase{key: "string", expect: "", expectOK: false})
	suite.assertStruct(ctx, testCase{key: "struct", expect: structValue{Name: "Kid", Age: 0}, expectOK: false})

	// If we can't marshal one of the types, propagate that error back.
	ctx = context.Background()
	ctx = metadata.WithValue(ctx, "nope", make(chan int, 10))
	_, err = metadata.ToJSON(ctx)
	suite.Require().Error(err, "Should return an error when value contains a type that can't be marshaled")
}

func (suite *ValuesSuite) assertString(ctx context.Context, c testCase) {
	var out string
	ok := metadata.Value(ctx, c.key, &out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().Equal(c.expect, out)
}

func (suite *ValuesSuite) assertStringPtr(ctx context.Context, c testCase) {
	var out = new(string)
	ok := metadata.Value(ctx, c.key, out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func (suite *ValuesSuite) assertInt(ctx context.Context, c testCase) {
	var out int
	ok := metadata.Value(ctx, c.key, &out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func (suite *ValuesSuite) assertIntPtr(ctx context.Context, c testCase) {
	var out = new(int)
	ok := metadata.Value(ctx, c.key, out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func (suite *ValuesSuite) assertBool(ctx context.Context, c testCase) {
	var out bool
	ok := metadata.Value(ctx, c.key, &out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func (suite *ValuesSuite) assertBoolPtr(ctx context.Context, c testCase) {
	var out = new(bool)
	ok := metadata.Value(ctx, c.key, out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func (suite *ValuesSuite) assertStruct(ctx context.Context, c testCase) {
	var out structValue
	ok := metadata.Value(ctx, c.key, &out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func (suite *ValuesSuite) assertStructPtr(ctx context.Context, c testCase) {
	var out = new(structValue)
	ok := metadata.Value(ctx, c.key, out)
	suite.Require().Equal(c.expectOK, ok)
	suite.Require().EqualValues(c.expect, out)
}

func stringPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

type testCase struct {
	key      string
	expect   interface{}
	expectOK bool
}

type structValue struct {
	Name string
	Age  uint8
}

func TestValuesSuite(t *testing.T) {
	suite.Run(t, new(ValuesSuite))
}

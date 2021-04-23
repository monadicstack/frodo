// +build client

package generate_test

import (
	"testing"

	"github.com/monadicstack/frodo/example/names"
	namesrpc "github.com/monadicstack/frodo/example/names/gen"
	"github.com/monadicstack/frodo/internal/testext"
	"github.com/monadicstack/frodo/rpc/errors"
	"github.com/stretchr/testify/suite"
)

type JavaScriptClientSuite struct {
	testext.ExternalClientSuite
}

func (suite *JavaScriptClientSuite) SetupTest() {
	handler := names.NameServiceHandler{}
	gateway := namesrpc.NewNameServiceGateway(&handler)
	suite.StartService(":9100", gateway)
}

func (suite *JavaScriptClientSuite) TearDownTest() {
	suite.StopService()
}

func (suite *JavaScriptClientSuite) Run(testName string, expectedLines int) testext.ClientTestResults {
	output, err := testext.RunClientTest("node", "testdata/js/run_client.js", testName)
	suite.Require().NoError(err, "Executing client runner should not give an error")
	suite.Require().Len(output, expectedLines)
	return output
}

// Ensures that we get a connection refused error when connecting to a not-running server.
func (suite *JavaScriptClientSuite) TestNotConnected() {
	assert := suite.Require()
	output := suite.Run("NotConnected", 1)

	fail0 := errors.RPCError{}
	suite.ExpectFail(output, 0, &fail0, func() {
		assert.Contains(fail0.Message, "ECONNREFUSED")
	})
}

// Ensures that the client fails gracefully if you injected a garbage 'fetch' implementation.
func (suite *JavaScriptClientSuite) TestBadFetch() {
	output := suite.Run("BadFetch", 1)
	suite.ExpectFail(output, 0, &errors.RPCError{})
}

// Ensures that all of our service functions succeed with valid inputs to the remote service.
func (suite *JavaScriptClientSuite) TestSuccess() {
	assert := suite.Require()
	output := suite.Run("Success", 5)

	res0 := names.SplitResponse{}
	suite.ExpectPass(output, 0, &res0, func() {
		assert.Equal("Jeff", res0.FirstName)
		assert.Equal("Lebowski", res0.LastName)
	})

	res1 := names.FirstNameResponse{}
	suite.ExpectPass(output, 1, &res1, func() {
		assert.Equal("Jeff", res1.FirstName)
	})

	res2 := names.LastNameResponse{}
	suite.ExpectPass(output, 2, &res2, func() {
		assert.Equal("Lebowski", res2.LastName)
	})

	res3 := names.SortNameResponse{}
	suite.ExpectPass(output, 3, &res3, func() {
		assert.Equal("lebowski, jeff", res3.SortName)
	})

	res4 := names.SortNameResponse{}
	suite.ExpectPass(output, 4, &res4, func() {
		assert.Equal("dude", res4.SortName)
	})
}

// Ensures that validation failures are properly propagated from the server.
func (suite *JavaScriptClientSuite) TestValidationFailure() {
	output := suite.Run("ValidationFailure", 8)

	suite.ExpectFailStatus(output, 0, 400)
	suite.ExpectFailStatus(output, 1, 400)
	suite.ExpectFailStatus(output, 2, 400)
	suite.ExpectFailStatus(output, 3, 400)
	suite.ExpectFailStatus(output, 4, 400)
	suite.ExpectFailStatus(output, 5, 400)
	suite.ExpectFailStatus(output, 6, 400)
	suite.ExpectFailStatus(output, 7, 400)
}

// Ensures that calls fail with a 403 if you have a bad authorization value on the entire client.
func (suite *JavaScriptClientSuite) TestAuthFailureClient() {
	output := suite.Run("AuthFailureClient", 4)

	suite.ExpectFailStatus(output, 0, 403)
	suite.ExpectFailStatus(output, 1, 403)
	suite.ExpectFailStatus(output, 2, 403)
	suite.ExpectFailStatus(output, 3, 403)
}

// Ensures that calls fail with a 403 if you have a bad authorization value on individual calls.
func (suite *JavaScriptClientSuite) TestAuthFailureCall() {
	output := suite.Run("AuthFailureCall", 4)

	suite.ExpectFailStatus(output, 0, 403)
	suite.ExpectFailStatus(output, 1, 403)
	suite.ExpectFailStatus(output, 2, 403)
	suite.ExpectFailStatus(output, 3, 403)
}

// Ensures that you can set a bad authorization on the client but valid auth on individual calls and
// it will work as expected.
func (suite *JavaScriptClientSuite) TestAuthFailureCallOverride() {
	assert := suite.Require()
	output := suite.Run("AuthFailureCallOverride", 4)

	res0 := names.SplitResponse{}
	suite.ExpectPass(output, 0, &res0, func() {
		assert.Equal("Dude", res0.FirstName)
	})

	res1 := names.FirstNameResponse{}
	suite.ExpectPass(output, 1, &res1, func() {
		assert.Equal("Dude", res1.FirstName)
	})

	res2 := names.LastNameResponse{}
	suite.ExpectPass(output, 2, &res2, func() {
		assert.Equal("", res2.LastName)
	})

	res3 := names.SortNameResponse{}
	suite.ExpectPass(output, 3, &res3, func() {
		assert.Equal("dude", res3.SortName)
	})
}

func TestJavaScriptClientSuite(t *testing.T) {
	suite.Run(t, new(JavaScriptClientSuite))
}

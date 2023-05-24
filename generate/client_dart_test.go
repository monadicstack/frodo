//go:build client
// +build client

package generate_test

import (
	"testing"

	"github.com/davidrenne/frodo/example/names"
	namesrpc "github.com/davidrenne/frodo/example/names/gen"
	"github.com/davidrenne/frodo/internal/testext"
	"github.com/davidrenne/frodo/rpc/errors"
	"github.com/stretchr/testify/suite"
)

type DartClientSuite struct {
	testext.ExternalClientSuite
}

func (suite *DartClientSuite) SetupTest() {
	handler := names.NameServiceHandler{}
	gateway := namesrpc.NewNameServiceGateway(&handler)
	suite.StartService(":9100", gateway)
}

func (suite *DartClientSuite) TearDownTest() {
	suite.StopService()
}

func (suite *DartClientSuite) Run(testName string, expectedLines int) testext.ClientTestResults {
	output, err := testext.RunClientTest("dart", "testdata/dart/run_client.dart", testName)
	suite.Require().NoError(err, "Executing client runner should not give an error: %v", err)
	suite.Require().Len(output, expectedLines)
	return output
}

// Ensures that we get a connection refused error when connecting to a not-running server.
func (suite *DartClientSuite) TestNotConnected() {
	assert := suite.Require()
	output := suite.Run("NotConnected", 2)

	fail := errors.RPCError{}
	suite.ExpectFail(output, 0, &fail, func() {
		assert.Contains(fail.Message, "Connection refused")
	})

	fail = errors.RPCError{}
	suite.ExpectFail(output, 1, &fail, func() {
		assert.Contains(fail.Message, "Connection refused")
	})
}

// Ensures that all of our service functions succeed with valid inputs to the remote service.
func (suite *DartClientSuite) TestSuccess() {
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

func (suite *DartClientSuite) TestSuccessRaw() {
	assert := suite.Require()
	output := suite.Run("SuccessRaw", 1)

	res := RawDartResult{}
	suite.ExpectPass(output, 0, &res, func() {
		assert.Equal("first,last\nJeff,Lebowski", res.Content)
		assert.Equal("application/octet-stream", res.ContentType)
		assert.Equal("", res.ContentFileName)
	})
}

func (suite *DartClientSuite) TestSuccessRawHeaders() {
	assert := suite.Require()
	output := suite.Run("SuccessRawHeaders", 3)

	res := RawNodeResult{}
	suite.ExpectPass(output, 0, &res, func() {
		assert.Equal("first,last\nJeff,Lebowski", res.Content)
		assert.Equal("text/csv", res.ContentType)
		assert.Equal("name.csv", res.ContentFileName)
	})

	res = RawNodeResult{}
	suite.ExpectPass(output, 1, &res, func() {
		assert.Equal("first,last\nJeff,Lebowski", res.Content)
		assert.Equal("text/txt", res.ContentType)
		assert.Equal("name.txt", res.ContentFileName)
	})

	res = RawNodeResult{}
	suite.ExpectPass(output, 2, &res, func() {
		assert.Equal("first,last\nJeff,Lebowski", res.Content)
		assert.Equal(`text/t"x"t`, res.ContentType)
		assert.Equal(`name.t"x"t`, res.ContentFileName)
	})
}

// Ensures that validation failures are properly propagated from the server.
func (suite *DartClientSuite) TestValidationFailure() {
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
func (suite *DartClientSuite) TestAuthFailureClient() {
	output := suite.Run("AuthFailureClient", 4)

	suite.ExpectFailStatus(output, 0, 403)
	suite.ExpectFailStatus(output, 1, 403)
	suite.ExpectFailStatus(output, 2, 403)
	suite.ExpectFailStatus(output, 3, 403)
}

// Ensures that calls fail with a 403 if you have a bad authorization value on individual calls.
func (suite *DartClientSuite) TestAuthFailureCall() {
	output := suite.Run("AuthFailureCall", 4)

	suite.ExpectFailStatus(output, 0, 403)
	suite.ExpectFailStatus(output, 1, 403)
	suite.ExpectFailStatus(output, 2, 403)
	suite.ExpectFailStatus(output, 3, 403)
}

// Ensures that you can set a bad authorization on the client but valid auth on individual calls and
// it will work as expected.
func (suite *DartClientSuite) TestAuthFailureCallOverride() {
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

func TestDartClientSuite(t *testing.T) {
	suite.Run(t, new(DartClientSuite))
}

// RawDartResult matches the data structure of the Node/JS object returned by service
// functions that result in "raw" byte responses.
type RawDartResult struct {
	// Content contains the raw byte content output by the service call.
	Content string
	// ContentType contains the captured "Content-Type" header data.
	ContentType string
	// ContentFileName contains the captured file name from the "Content-Disposition" header data.
	ContentFileName string
}

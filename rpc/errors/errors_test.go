package errors_test

import (
	"testing"

	"github.com/monadicstack/frodo/rpc/errors"
	"github.com/stretchr/testify/suite"
)

type ErrorsSuite struct {
	suite.Suite
}

func (suite *ErrorsSuite) TestNew() {
	suite.assertError(errors.New(100, "foo"), 100, "foo")
	suite.assertError(errors.New(100, "%s", "foo"), 100, "foo")
	suite.assertError(errors.New(100, "foo %s %v", "bar", 99), 100, "foo bar 99")
}

// Since we don't return an error argument for... creating an error... we have no
// problems with you providing non-HTTP standard errors. Maybe you're doing your own
// error status mapping; who's to say.
func (suite *ErrorsSuite) TestNew_wonkyStatus() {
	suite.assertError(errors.New(0, ""), 0, "")
	suite.assertError(errors.New(-42, ""), -42, "")
	suite.assertError(errors.New(9999, ""), 9999, "")
}

func (suite *ErrorsSuite) TestInternalServerError() {
	expectedStatus := 500
	suite.assertError(errors.InternalServerError("foo"), expectedStatus, "foo")
	suite.assertError(errors.InternalServerError("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.InternalServerError("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestBadRequest() {
	expectedStatus := 400
	suite.assertError(errors.BadRequest("foo"), expectedStatus, "foo")
	suite.assertError(errors.BadRequest("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.BadRequest("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestBadCredentials() {
	expectedStatus := 401
	suite.assertError(errors.BadCredentials("foo"), expectedStatus, "foo")
	suite.assertError(errors.BadCredentials("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.BadCredentials("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestPermissionDenied() {
	expectedStatus := 403
	suite.assertError(errors.PermissionDenied("foo"), expectedStatus, "foo")
	suite.assertError(errors.PermissionDenied("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.PermissionDenied("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestNotFound() {
	expectedStatus := 404
	suite.assertError(errors.NotFound("foo"), expectedStatus, "foo")
	suite.assertError(errors.NotFound("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.NotFound("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestAlreadyExists() {
	expectedStatus := 409
	suite.assertError(errors.AlreadyExists("foo"), expectedStatus, "foo")
	suite.assertError(errors.AlreadyExists("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.AlreadyExists("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestTimeout() {
	expectedStatus := 408
	suite.assertError(errors.Timeout("foo"), expectedStatus, "foo")
	suite.assertError(errors.Timeout("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Timeout("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestThrottled() {
	expectedStatus := 429
	suite.assertError(errors.Throttled("foo"), expectedStatus, "foo")
	suite.assertError(errors.Throttled("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Throttled("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestUnavailable() {
	expectedStatus := 503
	suite.assertError(errors.Unavailable("foo"), expectedStatus, "foo")
	suite.assertError(errors.Unavailable("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Unavailable("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

// assertError checks that both the status and message of the resulting 'err' are what we expect.
func (suite *ErrorsSuite) assertError(err errors.RPCError, expectedStatus int, expectedMessage string) {
	suite.Require().Equal(expectedStatus, err.Status())
	suite.Require().Equal(expectedMessage, err.Error())
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsSuite))
}

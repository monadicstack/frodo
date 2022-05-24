// +build unit

package errors_test

import (
	"fmt"
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

func (suite *ErrorsSuite) TestStatus() {
	suite.Equal(400, errors.Status(errors.BadRequest("")))
	suite.Equal(400, errors.Status(errors.New(400, "")))

	suite.Equal(503, errors.Status(errors.Unavailable("")))
	suite.Equal(503, errors.Status(errors.New(503, "")))

	suite.Equal(500, errors.Status(errors.InternalServerError("")))
	suite.Equal(500, errors.Status(errors.Unexpected("")))
	suite.Equal(500, errors.Status(errors.New(500, "")))

	// No status info at all
	suite.Equal(500, errors.Status(fmt.Errorf("hello")))

	// Non-RPCError examples that do have status values
	suite.Equal(500, errors.Status(errWithCode{code: 500}))
	suite.Equal(503, errors.Status(errWithCode{code: 503}))
	suite.Equal(404, errors.Status(errWithCode{code: 404}))
	suite.Equal(500, errors.Status(errWithStatusCode{statusCode: 500}))
	suite.Equal(503, errors.Status(errWithStatusCode{statusCode: 503}))
	suite.Equal(404, errors.Status(errWithStatusCode{statusCode: 404}))
}

func (suite *ErrorsSuite) TestUnexpected() {
	expectedStatus := 500
	suite.assertError(errors.Unexpected("foo"), expectedStatus, "foo")
	suite.assertError(errors.Unexpected("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Unexpected("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestUnexpectedASDF() {
	var err error
	switch {
	case errors.IsNotFound(err):
		// do stuff
	case err != nil:
		// do stuff
	default:
		// do stuff
	}
}

func (suite *ErrorsSuite) TestIsUnexpected() {
	suite.True(errors.IsUnexpected(errors.Unexpected("foo")))
	suite.True(errors.IsUnexpected(errors.InternalServerError("foo")))
	suite.True(errors.IsUnexpected(fmt.Errorf("no status present still 500")))
	suite.False(errors.IsUnexpected(errors.NotFound("")))
	suite.False(errors.IsUnexpected(errors.BadCredentials("")))

	// Non-RPCError examples
	suite.True(errors.IsUnexpected(errWithCode{code: 500}))
	suite.False(errors.IsUnexpected(errWithCode{code: 400}))
	suite.True(errors.IsUnexpected(errWithStatusCode{statusCode: 500}))
	suite.False(errors.IsUnexpected(errWithStatusCode{statusCode: 400}))
}

func (suite *ErrorsSuite) TestInternalServerError() {
	expectedStatus := 500
	suite.assertError(errors.InternalServerError("foo"), expectedStatus, "foo")
	suite.assertError(errors.InternalServerError("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.InternalServerError("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsInternalServiceError() {
	suite.True(errors.IsInternalServiceError(errors.Unexpected("foo")))
	suite.True(errors.IsInternalServiceError(errors.InternalServerError("foo")))
	suite.True(errors.IsInternalServiceError(fmt.Errorf("no status present still 500")))
	suite.False(errors.IsInternalServiceError(errors.NotFound("")))
	suite.False(errors.IsInternalServiceError(errors.BadCredentials("")))

	// Non-RPCError examples
	suite.True(errors.IsInternalServiceError(errWithCode{code: 500}))
	suite.False(errors.IsInternalServiceError(errWithCode{code: 400}))
	suite.True(errors.IsInternalServiceError(errWithStatusCode{statusCode: 500}))
	suite.False(errors.IsInternalServiceError(errWithStatusCode{statusCode: 400}))
}

func (suite *ErrorsSuite) TestBadRequest() {
	expectedStatus := 400
	suite.assertError(errors.BadRequest("foo"), expectedStatus, "foo")
	suite.assertError(errors.BadRequest("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.BadRequest("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsBadRequest() {
	suite.True(errors.IsBadRequest(errors.BadRequest("")))
	suite.False(errors.IsBadRequest(errors.Unexpected("")))
	suite.False(errors.IsBadRequest(errors.NotFound("")))

	// Non-RPCError examples
	suite.True(errors.IsBadRequest(errWithCode{code: 400}))
	suite.False(errors.IsBadRequest(errWithCode{code: 403}))
	suite.True(errors.IsBadRequest(errWithStatusCode{statusCode: 400}))
	suite.False(errors.IsBadRequest(errWithStatusCode{statusCode: 403}))
}

func (suite *ErrorsSuite) TestBadCredentials() {
	expectedStatus := 401
	suite.assertError(errors.BadCredentials("foo"), expectedStatus, "foo")
	suite.assertError(errors.BadCredentials("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.BadCredentials("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsBadCredentials() {
	suite.True(errors.IsBadCredentials(errors.BadCredentials("")))
	suite.False(errors.IsBadCredentials(errors.Unexpected("")))
	suite.False(errors.IsBadCredentials(errors.NotFound("")))

	// Non-RPCError examples
	suite.True(errors.IsBadCredentials(errWithCode{code: 401}))
	suite.False(errors.IsBadCredentials(errWithCode{code: 403}))
	suite.True(errors.IsBadCredentials(errWithStatusCode{statusCode: 401}))
	suite.False(errors.IsBadCredentials(errWithStatusCode{statusCode: 403}))
}

func (suite *ErrorsSuite) TestPermissionDenied() {
	expectedStatus := 403
	suite.assertError(errors.PermissionDenied("foo"), expectedStatus, "foo")
	suite.assertError(errors.PermissionDenied("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.PermissionDenied("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsPermissionDenied() {
	suite.True(errors.IsPermissionDenied(errors.PermissionDenied("")))
	suite.False(errors.IsPermissionDenied(errors.Unexpected("")))
	suite.False(errors.IsPermissionDenied(errors.NotFound("")))

	// Non-RPCError examples
	suite.True(errors.IsPermissionDenied(errWithCode{code: 403}))
	suite.False(errors.IsPermissionDenied(errWithCode{code: 404}))
	suite.True(errors.IsPermissionDenied(errWithStatusCode{statusCode: 403}))
	suite.False(errors.IsPermissionDenied(errWithStatusCode{statusCode: 404}))
}

func (suite *ErrorsSuite) TestNotFound() {
	expectedStatus := 404
	suite.assertError(errors.NotFound("foo"), expectedStatus, "foo")
	suite.assertError(errors.NotFound("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.NotFound("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsNotFound() {
	suite.True(errors.IsNotFound(errors.NotFound("")))
	suite.False(errors.IsNotFound(errors.Unexpected("")))
	suite.False(errors.IsNotFound(errors.BadRequest("")))

	// Non-RPCError examples
	suite.True(errors.IsNotFound(errWithCode{code: 404}))
	suite.False(errors.IsNotFound(errWithCode{code: 401}))
	suite.True(errors.IsNotFound(errWithStatusCode{statusCode: 404}))
	suite.False(errors.IsNotFound(errWithStatusCode{statusCode: 401}))
}

func (suite *ErrorsSuite) TestAlreadyExists() {
	expectedStatus := 409
	suite.assertError(errors.AlreadyExists("foo"), expectedStatus, "foo")
	suite.assertError(errors.AlreadyExists("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.AlreadyExists("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsAlreadyExists() {
	suite.True(errors.IsAlreadyExists(errors.AlreadyExists("")))
	suite.False(errors.IsAlreadyExists(errors.Unexpected("")))
	suite.False(errors.IsAlreadyExists(errors.BadRequest("")))

	// Non-RPCError examples
	suite.True(errors.IsAlreadyExists(errWithCode{code: 409}))
	suite.False(errors.IsAlreadyExists(errWithCode{code: 401}))
	suite.True(errors.IsAlreadyExists(errWithStatusCode{statusCode: 409}))
	suite.False(errors.IsAlreadyExists(errWithStatusCode{statusCode: 401}))
}

func (suite *ErrorsSuite) TestTimeout() {
	expectedStatus := 408
	suite.assertError(errors.Timeout("foo"), expectedStatus, "foo")
	suite.assertError(errors.Timeout("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Timeout("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsTimeout() {
	suite.True(errors.IsTimeout(errors.Timeout("")))
	suite.False(errors.IsTimeout(errors.Unexpected("")))
	suite.False(errors.IsTimeout(errors.BadRequest("")))

	// Non-RPCError examples
	suite.True(errors.IsTimeout(errWithCode{code: 408}))
	suite.False(errors.IsTimeout(errWithCode{code: 401}))
	suite.True(errors.IsTimeout(errWithStatusCode{statusCode: 408}))
	suite.False(errors.IsTimeout(errWithStatusCode{statusCode: 401}))
}

func (suite *ErrorsSuite) TestThrottled() {
	expectedStatus := 429
	suite.assertError(errors.Throttled("foo"), expectedStatus, "foo")
	suite.assertError(errors.Throttled("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Throttled("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsThrottled() {
	suite.True(errors.IsThrottled(errors.Throttled("")))
	suite.False(errors.IsThrottled(errors.Unexpected("")))
	suite.False(errors.IsThrottled(errors.BadRequest("")))

	// Non-RPCError examples
	suite.True(errors.IsThrottled(errWithCode{code: 429}))
	suite.False(errors.IsThrottled(errWithCode{code: 401}))
	suite.True(errors.IsThrottled(errWithStatusCode{statusCode: 429}))
	suite.False(errors.IsThrottled(errWithStatusCode{statusCode: 401}))
}

func (suite *ErrorsSuite) TestUnavailable() {
	expectedStatus := 503
	suite.assertError(errors.Unavailable("foo"), expectedStatus, "foo")
	suite.assertError(errors.Unavailable("%s", "foo"), expectedStatus, "foo")
	suite.assertError(errors.Unavailable("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *ErrorsSuite) TestIsUnavailable() {
	suite.True(errors.IsUnavailable(errors.Unavailable("")))
	suite.False(errors.IsUnavailable(errors.Unexpected("")))
	suite.False(errors.IsUnavailable(errors.BadRequest("")))

	// Non-RPCError examples
	suite.True(errors.IsUnavailable(errWithCode{code: 503}))
	suite.False(errors.IsUnavailable(errWithCode{code: 401}))
	suite.True(errors.IsUnavailable(errWithStatusCode{statusCode: 503}))
	suite.False(errors.IsUnavailable(errWithStatusCode{statusCode: 401}))
}

// assertError checks that both the status and message of the resulting 'err' are what we expect.
func (suite *ErrorsSuite) assertError(err errors.RPCError, expectedStatus int, expectedMessage string) {
	suite.Require().Equal(expectedStatus, err.Status())
	suite.Require().Equal(expectedMessage, err.Error())
}

func TestErrorsSuite(t *testing.T) {
	suite.Run(t, new(ErrorsSuite))
}

type errWithCode struct {
	code int
}

func (err errWithCode) Code() int {
	return err.code
}

func (err errWithCode) Error() string {
	return ""
}

type errWithStatusCode struct {
	statusCode int
}

func (err errWithStatusCode) StatusCode() int {
	return err.statusCode
}

func (err errWithStatusCode) Error() string {
	return ""
}

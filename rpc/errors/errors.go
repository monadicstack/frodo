// Package errors provides a curated list of standard types of failures you'll likely
// encounter when executing RPC handlers. There are almost 40 failure status codes in
// standard HTTP which leads to a lot of confusion about which to use in certain cases.
// To simplify this decision making, the errors package only exposes a dozen or so types
// of failures that we'll map to an appropriate HTTP/RPC status for you. They represent
// the most common types of failure you're likely to encounter. Should you need something
// beyond this, the New() function accepts any status you feel like generating.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// RPCError is an error that encodes a human-readable message as well as a
// status that corresponds to the closest HTTP status code. For instance if the failure
// was due to an inability to find a resource/record/etc, status would be 404.
//
// This type of error can be recognized by 'github.com/monadicstack/respond' in order to
// automatically send proper failures and statuses without extra lifting.
type RPCError struct {
	// HTTPStatus is the HTTP status code that most closely describes this error.
	HTTPStatus int `json:"status"`
	// Message is the human-readable error message.
	Message string `json:"message"`
}

// Error returns the underlying error message that describes this failure.
func (err RPCError) Error() string {
	return err.Message
}

// Status returns the most relevant HTTP status code to return for this error.
func (err RPCError) Status() int {
	return err.HTTPStatus
}

type errorWithStatus interface {
	Status() int
}

type errorWithStatusCode interface {
	StatusCode() int
}

type errorWithCode interface {
	Code() int
}

// New creates an error that maps directly to an HTTP status so if your RPC call results in
// this error, it will result in the same 'status' in your HTTP response. While you can do this
// for more obscure HTTP failure statuses like "payment required", it's typically a better idea
// to use the error functions BadRequest(), PermissionDenied(), etc as it provides proper status
// codes and results in more readable code.
func New(status int, messageFormat string, args ...interface{}) RPCError {
	return RPCError{
		HTTPStatus: status,
		Message:    fmt.Sprintf(messageFormat, args...),
	}
}

// Status looks for either a Status(), StatusCode(), or Code() method on the error to
// figure out the most appropriate HTTP status code for it. If the error doesn't have any of
// those methods then we'll just assume that it is a 500 error.
func Status(err error) int {
	var errStatus errorWithStatus
	if errors.As(err, &errStatus) {
		return errStatus.Status()
	}

	var errStatusCode errorWithStatusCode
	if errors.As(err, &errStatusCode) {
		return errStatusCode.StatusCode()
	}

	var errCode errorWithCode
	if errors.As(err, &errCode) {
		return errCode.Code()
	}

	return http.StatusInternalServerError
}

// Unexpected is a generic 500-style catch-all error for failures you don't know what to do with. This is
// exactly the same as calling InternalServerError(), just more concise in your code.
func Unexpected(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusInternalServerError, messageFormat, args...)
}

// IsUnexpected returns true if the underlying HTTP status code of 'err' is 500. This will be true for any
// error you created using the Unexpected() function.
func IsUnexpected(err error) bool {
	return Status(err) == http.StatusInternalServerError
}

// InternalServerError is a generic 500-style catch-all error for failures you don't know what to do with.
func InternalServerError(messageFormat string, args ...interface{}) RPCError {
	return Unexpected(messageFormat, args...)
}

// IsInternalServiceError returns true if the underlying HTTP status code of 'err' is 500. This will be true for any
// error you created using the InternalServiceError() function.
func IsInternalServiceError(err error) bool {
	return Status(err) == http.StatusInternalServerError
}

// BadRequest is a 400-style error that indicates that some aspect of the request was either ill-formed
// or failed validation. This could be an ill-formed function parameter, a bad HTTP body, etc.
func BadRequest(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusBadRequest, messageFormat, args...)
}

// IsBadRequest returns true if the underlying HTTP status code of 'err' is 400. This will be true for any
// error you created using the BadRequest() function.
func IsBadRequest(err error) bool {
	return Status(err) == http.StatusBadRequest
}

// BadCredentials is a 401-style error that indicates that the caller either didn't provide credentials
// when necessary or they did, but the credentials were invalid for some reason. This corresponds to the
// HTTP "unauthorized" status, but we prefer this name because this type of failure has nothing to do
// with authorization and it's more clear what aspect of the request has failed.
func BadCredentials(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusUnauthorized, messageFormat, args...)
}

// IsBadCredentials returns true if the underlying HTTP status code of 'err' is 401. This will be true for any
// error you created using the BadCredentials() function.
func IsBadCredentials(err error) bool {
	return Status(err) == http.StatusUnauthorized
}

// PermissionDenied is a 403-style error that indicates that the caller does not have rights/clearance
// to perform any part of the operation.
func PermissionDenied(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusForbidden, messageFormat, args...)
}

// IsPermissionDenied returns true if the underlying HTTP status code of 'err' is 403. This will be true for any
// error you created using the PermissionDenied() function.
func IsPermissionDenied(err error) bool {
	return Status(err) == http.StatusForbidden
}

// NotFound is a 404-style error that indicates that some record/resource could not be located.
func NotFound(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusNotFound, messageFormat, args...)
}

// IsNotFound returns true if the underlying HTTP status code of 'err' is 404. This will be true for any
// error you created using the NotFound() function.
func IsNotFound(err error) bool {
	return Status(err) == http.StatusNotFound
}

// Timeout is a 408-style error that indicates that some operation exceeded its allotted time/deadline.
func Timeout(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusRequestTimeout, messageFormat, args...)
}

// IsTimeout returns true if the underlying HTTP status code of 'err' is 408. This will be true for any
// error you created using the Timeout() function.
func IsTimeout(err error) bool {
	return Status(err) == http.StatusRequestTimeout
}

// AlreadyExists is a 409-style error that is used when attempting to create some record/resource, but
// there is already a duplicate instance in existence.
func AlreadyExists(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusConflict, messageFormat, args...)
}

// IsAlreadyExists returns true if the underlying HTTP status code of 'err' is 409. This will be true for any
// error you created using the AlreadyExists() function.
func IsAlreadyExists(err error) bool {
	return Status(err) == http.StatusConflict
}

// Throttled is a 429-style error that indicates that the caller has exceeded the number of requests,
// amount of resources, etc allowed over some time period. The failure should indicated to the caller
// that the failure is due to some throttle that prevented the operation from even occurring.
func Throttled(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusTooManyRequests, messageFormat, args...)
}

// IsThrottled returns true if the underlying HTTP status code of 'err' is 429. This will be true for any
// error you created using the Throttled() function.
func IsThrottled(err error) bool {
	return Status(err) == http.StatusTooManyRequests
}

// Unavailable is a 503-style error that indicates that some aspect of the server/service is unavailable.
// This could be something like DB connection failures, some third party service being down, etc.
func Unavailable(messageFormat string, args ...interface{}) RPCError {
	return New(http.StatusServiceUnavailable, messageFormat, args...)
}

// IsUnavailable returns true if the underlying HTTP status code of 'err' is 503. This will be true for any
// error you created using the Unavailable() function.
func IsUnavailable(err error) bool {
	return Status(err) == http.StatusServiceUnavailable
}

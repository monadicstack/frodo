// +build unit

package rpc_test

import (
	"net/http"
	"testing"

	"github.com/monadicstack/frodo/rpc"
	"github.com/stretchr/testify/suite"
)

/*
 * The chaining of middleware is actually well-exercised in the Client and Gateway tests, so there's no need
 * to duplicate those here. Additionally, most of the things required to test that are not exported so it would
 * be hard to test here anyway.
 */

type MiddlewareSuite struct {
	suite.Suite
}

// Ensures that a basic MiddlewareFunc can behave properly when used as a Middleware interface instance.
func (suite *MiddlewareSuite) TestServeHTTP() {
	values := []string{"", ""}
	mw := rpc.MiddlewareFunc(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		values[0] = "Hello"
		next(w, req)
	})
	mw.ServeHTTP(nil, nil, func(http.ResponseWriter, *http.Request) {
		values[1] = "World"
	})

	suite.Require().Equal("Hello", values[0], "Middleware.ServeHTTP should invoke the underlying function.")
	suite.Require().Equal("World", values[1], "Middleware.ServeHTTP should invoke the underlying function.")
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}

// +build client

package generate_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/monadicstack/frodo/example/basic/calc"
	calcrpc "github.com/monadicstack/frodo/example/basic/calc/gen"
	"github.com/monadicstack/frodo/rpc"
	"github.com/monadicstack/frodo/rpc/errors"
	"github.com/stretchr/testify/suite"
)

type GoClientSuite struct {
	suite.Suite
	server *http.Server
	client *calcrpc.CalculatorServiceClient
}

func (suite *GoClientSuite) SetupTest() {
	serviceHandler := calc.CalculatorServiceHandler{}
	gateway := calcrpc.NewCalculatorServiceGateway(&serviceHandler)
	suite.server = &http.Server{Addr: ":54242", Handler: gateway}
	go func() {
		_ = suite.server.ListenAndServe()
	}()

	suite.client = calcrpc.NewCalculatorServiceClient("http://localhost:54242")
}

func (suite *GoClientSuite) TearDownTest() {
	if suite.server != nil {
		_ = suite.server.Shutdown(context.Background())
	}
}

func (suite *GoClientSuite) TestMethodSuccess() {
	r := suite.Require()
	ctx := context.Background()

	add, err := suite.client.Add(ctx, &calc.AddRequest{A: 5, B: 3})
	r.NoError(err, "Adding two positive numbers should not result in an error")
	r.Equal(8, add.Result)

	sub, err := suite.client.Sub(ctx, &calc.SubRequest{A: 5, B: 3})
	r.NoError(err, "Subtracting two positive numbers should not result in an error")
	r.Equal(2, sub.Result)
}

func (suite *GoClientSuite) TestMethodFailure() {
	r := suite.Require()
	ctx := context.Background()

	_, err := suite.client.Add(ctx, &calc.AddRequest{A: 12344, B: 1})
	r.Error(err, "Adding up to 12345 should result in a 403 error")
	rpcErr, ok := err.(errors.RPCError)
	r.True(ok, "Client should convert all HTTP error status codes to RPCError instances")
	r.Equal(403, rpcErr.Status())

	_, err = suite.client.Sub(ctx, &calc.SubRequest{A: 3, B: 5})
	r.Error(err, "Subtraction doesn't allow negative results")
	rpcErr, ok = err.(errors.RPCError)
	r.True(ok, "Client should convert all HTTP error status codes to RPCError instances")
	r.Equal(400, rpcErr.Status())
}

func (suite *GoClientSuite) TestParamNilChecks() {
	r := suite.Require()
	ctx := context.Background()

	_, err := suite.client.Add(nil, &calc.AddRequest{A: 5, B: 3})
	r.Error(err, "Should fail if context is nil")

	_, err = suite.client.Add(ctx, nil)
	r.Error(err, "Should fail if method request parameter is nil")

	_, err = suite.client.Sub(nil, &calc.SubRequest{A: 5, B: 3})
	r.Error(err, "Should fail if context is nil")

	_, err = suite.client.Sub(ctx, nil)
	r.Error(err, "Should fail if method request parameter is nil")
}

func (suite *GoClientSuite) TestInvalidEndpoint() {
	r := suite.Require()
	ctx := context.Background()

	// Port is different than the gateway in SetupTest
	client := calcrpc.NewCalculatorServiceClient("http://localhost:55555")

	_, err := client.Add(ctx, &calc.AddRequest{A: 5, B: 3})
	r.Error(err, "Should fail if unable to connect to gateway endpoint")
	r.Contains(err.Error(), "rpc: round trip error")

	_, err = client.Sub(ctx, &calc.SubRequest{A: 5, B: 3})
	r.Error(err, "Should fail if unable to connect to gateway endpoint")
	r.Contains(err.Error(), "rpc: round trip error")
}

func (suite *GoClientSuite) TestMiddleware() {
	r := suite.Require()
	ctx := context.Background()
	var values []string

	client := calcrpc.NewCalculatorServiceClient("http://localhost:54242", rpc.WithClientMiddleware(
		func(request *http.Request, next rpc.RoundTripperFunc) (*http.Response, error) {
			values = append(values, "a")
			res, err := next(request)
			values = append(values, "d")
			return res, err
		},
		func(request *http.Request, next rpc.RoundTripperFunc) (*http.Response, error) {
			values = append(values, "b")
			res, err := next(request)
			values = append(values, "c")
			return res, err
		},
	))

	_, err := client.Add(ctx, &calc.AddRequest{A: 5, B: 3})
	r.NoError(err)
	r.Equal([]string{"a", "b", "c", "d"}, values)

	_, err = client.Sub(ctx, &calc.SubRequest{A: 5, B: 3})
	r.NoError(err)
	r.Equal([]string{"a", "b", "c", "d", "a", "b", "c", "d"}, values)
}

func TestGoClientSuite(t *testing.T) {
	suite.Run(t, new(GoClientSuite))
}

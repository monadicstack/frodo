//go:build client
// +build client

package generate_test

import (
	"context"
	stderrors "errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/davidrenne/frodo/example/names"
	namesrpc "github.com/davidrenne/frodo/example/names/gen"
	"github.com/davidrenne/frodo/rpc"
	"github.com/davidrenne/frodo/rpc/authorization"
	"github.com/davidrenne/frodo/rpc/errors"
	"github.com/stretchr/testify/suite"
)

type GoClientSuite struct {
	suite.Suite
	server *http.Server
	client *namesrpc.NameServiceClient
}

func (suite *GoClientSuite) SetupTest() {
	serviceHandler := names.NameServiceHandler{}
	gateway := namesrpc.NewNameServiceGateway(&serviceHandler)
	suite.server = &http.Server{Addr: ":54242", Handler: gateway}
	go func() {
		_ = suite.server.ListenAndServe()
	}()

	suite.client = namesrpc.NewNameServiceClient("http://localhost:54242")
}

func (suite *GoClientSuite) TearDownTest() {
	if suite.server != nil {
		_ = suite.server.Shutdown(context.Background())
	}
}

func (suite *GoClientSuite) errorStatus(err error) int {
	var rpcErr errors.RPCError
	if stderrors.As(err, &rpcErr) {
		return rpcErr.Status()
	}
	return 0
}

// Ensures that we capture a "connection refused" error if we attempt to connect to a bad address for the service
// or it's not responding on that address.
func (suite *GoClientSuite) TestNotConnected() {
	r := suite.Require()
	ctx := context.Background()
	badClient := namesrpc.NewNameServiceClient("http://localhost:55555")

	_, err := badClient.Split(ctx, &names.SplitRequest{Name: "Jeff Lebowski"})
	r.Error(err, "Calls should not succeed if client can't connect to address")
	r.Contains(err.Error(), "connection refused")

	_, err = badClient.Download(ctx, &names.DownloadRequest{Name: "Jeff Lebowski"})
	r.Error(err, "Calls should not succeed if client can't connect to address")
	r.Contains(err.Error(), "connection refused")
}

// Ensures that JSON-based requests succeed if nothing goes wrong.
func (suite *GoClientSuite) TestSuccess() {
	r := suite.Require()
	ctx := context.Background()

	split, err := suite.client.Split(ctx, &names.SplitRequest{Name: "Jeff Lebowski"})
	r.NoError(err, "Successful calls should not result in an error")
	r.Equal("Jeff", split.FirstName)
	r.Equal("Lebowski", split.LastName)

	first, err := suite.client.FirstName(ctx, &names.FirstNameRequest{Name: "Jeff Lebowski"})
	r.NoError(err, "Successful calls should not result in an error")
	r.Equal("Jeff", first.FirstName)

	last, err := suite.client.LastName(ctx, &names.LastNameRequest{Name: "Jeff Lebowski"})
	r.NoError(err, "Successful calls should not result in an error")
	r.Equal("Lebowski", last.LastName)

	sort1, err := suite.client.SortName(ctx, &names.SortNameRequest{Name: "Jeff Lebowski"})
	r.NoError(err, "Successful calls should not result in an error")
	r.Equal("lebowski, jeff", sort1.SortName)

	sort2, err := suite.client.SortName(ctx, &names.SortNameRequest{Name: "Dude"})
	r.NoError(err, "Successful calls should not result in an error")
	r.Equal("dude", sort2.SortName)
}

// Ensures that explicit error status codes are preserved when the service returns strongly-coded errors.
func (suite *GoClientSuite) TestValidationFailure() {
	r := suite.Require()
	ctx := context.Background()

	assertError := func(_ interface{}, err error) {
		r.Error(err, "When service returns an error, client should propagate the error")
		r.Equal(400, suite.errorStatus(err), "Service errors should maintain status code")
	}

	assertError(suite.client.Split(ctx, &names.SplitRequest{Name: ""}))
	assertError(suite.client.FirstName(ctx, &names.FirstNameRequest{Name: ""}))
	assertError(suite.client.LastName(ctx, &names.LastNameRequest{Name: ""}))
	assertError(suite.client.SortName(ctx, &names.SortNameRequest{Name: ""}))

	// Frodo always treats errors as JSON even if the success response type is a "raw" response.
	assertError(suite.client.Download(ctx, &names.DownloadRequest{Name: ""}))
	assertError(suite.client.DownloadExt(ctx, &names.DownloadExtRequest{Name: ""}))
}

// Ensure that the client propagates 403-style errors returned by the service when it rejects the authorization
// value we supply w/ the context.
func (suite *GoClientSuite) TestAuthFailureCall() {
	r := suite.Require()
	ctx := context.Background()
	ctx = authorization.WithHeader(ctx, authorization.New("Donny"))

	assertError := func(_ interface{}, err error) {
		r.Error(err, "Calls should not succeed if context contained bad authorization credentials")
		r.Contains(err.Error(), "out of your element", "Authorization error should contain message from the service")
	}

	assertError(suite.client.Split(ctx, &names.SplitRequest{Name: "Jeff Lebowski"}))
	assertError(suite.client.FirstName(ctx, &names.FirstNameRequest{Name: "Jeff Lebowski"}))
	assertError(suite.client.LastName(ctx, &names.LastNameRequest{Name: "Jeff Lebowski"}))
	assertError(suite.client.SortName(ctx, &names.SortNameRequest{Name: "Jeff Lebowski"}))
}

// Ensures that we don't fail w/ a 403 if the service accepts our authorization credentials.
func (suite *GoClientSuite) TestAuthSuccessCall() {
	r := suite.Require()
	ctx := context.Background()
	ctx = authorization.WithHeader(ctx, authorization.New("Maude"))

	assertSuccess := func(_ interface{}, err error) {
		r.NoError(err, "Calls should succeed if service accepted the context authorization")
	}

	assertSuccess(suite.client.Split(ctx, &names.SplitRequest{Name: "Jeff Lebowski"}))
	assertSuccess(suite.client.FirstName(ctx, &names.FirstNameRequest{Name: "Jeff Lebowski"}))
	assertSuccess(suite.client.LastName(ctx, &names.LastNameRequest{Name: "Jeff Lebowski"}))
	assertSuccess(suite.client.SortName(ctx, &names.SortNameRequest{Name: "Jeff Lebowski"}))
}

func (suite *GoClientSuite) TestParamNilChecks() {
	r := suite.Require()
	ctx := context.Background()

	assertNilContextError := func(_ interface{}, err error) {
		r.Error(err, "Should fail if context is nil")
		r.Equal(0, suite.errorStatus(err), "Error should have 0 status when context is nil")
	}
	assertNilParamError := func(_ interface{}, err error) {
		r.Error(err, "Should fail if method request parameter is nil")
		r.Equal(0, suite.errorStatus(err), "Error should have 0 status when request parameter is nil")
	}

	assertNilContextError(suite.client.Split(nil, &names.SplitRequest{Name: "Dude"}))
	assertNilParamError(suite.client.Split(ctx, nil))

	assertNilContextError(suite.client.FirstName(nil, &names.FirstNameRequest{Name: "Dude"}))
	assertNilParamError(suite.client.FirstName(ctx, nil))

	assertNilContextError(suite.client.LastName(nil, &names.LastNameRequest{Name: "Dude"}))
	assertNilParamError(suite.client.LastName(ctx, nil))

	assertNilContextError(suite.client.SortName(nil, &names.SortNameRequest{Name: "Dude"}))
	assertNilParamError(suite.client.SortName(ctx, nil))

	assertNilContextError(suite.client.Download(nil, &names.DownloadRequest{Name: "Dude"}))
	assertNilParamError(suite.client.Download(ctx, nil))

	assertNilContextError(suite.client.DownloadExt(nil, &names.DownloadExtRequest{Name: "Dude"}))
	assertNilParamError(suite.client.DownloadExt(ctx, nil))
}

// Ensures that a response that implements ContentReader/Writer is properly populated w/ raw data rather
// than using JSON serialization.
func (suite *GoClientSuite) TestRaw() {
	r := suite.Require()
	ctx := context.Background()

	res, err := suite.client.Download(ctx, &names.DownloadRequest{Name: "Jeff Lebowski"})
	r.NoError(err, "Raw calls should succeed w/o failure")
	data, _ := ioutil.ReadAll(res.Content())
	r.Equal("first,last\nJeff,Lebowski", string(data))
}

// Ensures that a raw response will capture the content type and disposition file name if you implement
// the correct interfaces.
func (suite *GoClientSuite) TestRaw_withHeaders() {
	r := suite.Require()
	ctx := context.Background()

	res, err := suite.client.DownloadExt(ctx, &names.DownloadExtRequest{Name: "Jeff Lebowski", Ext: "csv"})
	r.NoError(err, "Raw calls should succeed w/o failure")
	data, _ := ioutil.ReadAll(res.Content())
	r.Equal("first,last\nJeff,Lebowski", string(data))
	r.Equal("text/csv", res.ContentType())
	r.Equal("name.csv", res.ContentFileName())

	res, err = suite.client.DownloadExt(ctx, &names.DownloadExtRequest{Name: "Walter Sobchak", Ext: "txt"})
	r.NoError(err, "Raw calls should succeed w/o failure")
	data, _ = ioutil.ReadAll(res.Content())
	r.Equal("first,last\nWalter,Sobchak", string(data))
	r.Equal("text/txt", res.ContentType())
	r.Equal("name.txt", res.ContentFileName())

	res, err = suite.client.DownloadExt(ctx, &names.DownloadExtRequest{Name: "Walter Sobchak", Ext: `t"x"t`})
	r.NoError(err, "Raw calls should succeed w/o failure")
	data, _ = ioutil.ReadAll(res.Content())
	r.Equal("first,last\nWalter,Sobchak", string(data))
	r.Equal(`text/t"x"t`, res.ContentType())
	r.Equal(`name.t"x"t`, res.ContentFileName())
}

// Ensures that middleware functions work properly when calling JSON-based methods on the service.
func (suite *GoClientSuite) TestMiddleware() {
	r := suite.Require()
	ctx := context.Background()
	var values []string

	client := namesrpc.NewNameServiceClient("http://localhost:54242", rpc.WithClientMiddleware(
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

	first, err := client.FirstName(ctx, &names.FirstNameRequest{Name: "Jeff Lebowski"})
	r.NoError(err)
	r.Equal("Jeff", first.FirstName)
	r.Equal([]string{"a", "b", "c", "d"}, values)
}

// Ensures that middleware functions work when calling raw-content functions on the service.
func (suite *GoClientSuite) TestMiddleware_raw() {
	r := suite.Require()
	ctx := context.Background()
	var values []string

	client := namesrpc.NewNameServiceClient("http://localhost:54242", rpc.WithClientMiddleware(
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

	download, err := client.Download(ctx, &names.DownloadRequest{Name: "Jeff Lebowski"})
	r.NoError(err)
	data, _ := ioutil.ReadAll(download.Content())
	r.Equal("first,last\nJeff,Lebowski", string(data))
	r.Equal([]string{"a", "b", "c", "d"}, values)
}

func TestGoClientSuite(t *testing.T) {
	suite.Run(t, new(GoClientSuite))
}

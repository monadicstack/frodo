package rpc_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/monadicstack/frodo/rpc"
	"github.com/stretchr/testify/suite"
)

type GatewaySuite struct {
	suite.Suite
	HTTPClient *http.Client
}

func (suite *GatewaySuite) SetupTest() {
	timeout := 1 * time.Second
	suite.HTTPClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: timeout}).DialContext,
			TLSHandshakeTimeout: timeout,
		},
	}
}

func (suite *GatewaySuite) TestNewGateway() {
	gateway := rpc.NewGateway()
	suite.Require().NotNil(gateway, "Gateway should be non-nil with no special options")
	suite.Require().NotNil(gateway.Binder, "Gateway should have binder by default")
	suite.Require().Equal("", gateway.PathPrefix, "Gateway should not have a path prefix by default")

	gateway = rpc.NewGateway(
		func(g *rpc.Gateway) { g.Binder = nil },
		func(g *rpc.Gateway) { g.Name = "Foo" },
		func(g *rpc.Gateway) { g.PathPrefix = "/fart" },
		func(g *rpc.Gateway) { g.PathPrefix = "/fart/again" },
		func(g *rpc.Gateway) { g.Name = "Bar" },
	)
	suite.Require().NotNil(gateway, "Gateway should be non-nil when given functional options")
	suite.Require().Nil(gateway.Binder, "Gateway should have functional options applied in order")
	suite.Require().Equal("/fart/again", gateway.PathPrefix, "Gateway should have functional options applied in order")
	suite.Require().Equal("Bar", gateway.Name, "Gateway should have functional options applied in order")
}

// Ensures that we respond with 404 to some otherwise common routes "GET /", "GET /ServiceName", "POST /ServiceName"
// if you have not registered any endpoints.
func (suite *GatewaySuite) TestNoRoutes() {
	server := httptest.NewServer(rpc.NewGateway(
		func(g *rpc.Gateway) { g.Name = "FooService" },
	))

	res, err := suite.HTTPClient.Get(server.URL)
	suite.Require().NoError(err)
	defer res.Body.Close()
	suite.Require().Equal(404, res.StatusCode, "Should not have any routes in the API with no endpoints registered")

	res, err = suite.HTTPClient.Get(server.URL + "/FooService")
	suite.Require().NoError(err)
	defer res.Body.Close()
	suite.Require().Equal(404, res.StatusCode, "Should not have any routes in the API with no endpoints registered")

	res, err = suite.HTTPClient.Post(server.URL+"/FooService", "", http.NoBody)
	suite.Require().NoError(err)
	defer res.Body.Close()
	suite.Require().Equal(404, res.StatusCode, "Should not have any routes in the API with no endpoints registered")
}

// Ensures that when you ".Register()" endpoints that the proper routes AND their OPTIONS counterparts are there.
func (suite *GatewaySuite) TestRegister() {
	gateway := rpc.NewGateway(func(g *rpc.Gateway) { g.Name = "FooService" })
	gateway.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "FooService.Hello",
		ServiceName: "FooService",
		Name:        "Hello",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`"hello"`))
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "get",
		Path:        "/foo/:a/:b",
		ServiceName: "FooService",
		Name:        "Sum",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			params := httprouter.ParamsFromContext(req.Context())
			w.WriteHeader(202)
			_, _ = w.Write([]byte(fmt.Sprintf(`"%s + %s = ?"`, params.ByName("a"), params.ByName("b"))))
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "DELETE",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Delete",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(204)
			_, _ = w.Write([]byte(fmt.Sprintf(`delete`)))
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "PUT",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Put",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(fmt.Sprintf(`put`)))
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "PATCH",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Patch",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(fmt.Sprintf(`patch`)))
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "HEAD",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Head",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(fmt.Sprintf(`head`)))
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, result, err := suite.request(server, "GET", "/foo/42/43", "")
	suite.Require().NoError(err)
	suite.Require().Equal(202, status, "Gateway GET request not properly responding")
	suite.Require().Equal(`"42 + 43 = ?"`, result, "Gateway GET request not properly responding")

	status, result, err = suite.request(server, "GET", "/foo/98/99", "")
	suite.Require().NoError(err)
	suite.Require().Equal(202, status, "Gateway GET request not properly responding")
	suite.Require().Equal(`"98 + 99 = ?"`, result, "Gateway GET request not properly responding")

	status, result, err = suite.request(server, "POST", "/FooService.Hello", `{"Foo":1}`)
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway POST request not properly responding")
	suite.Require().Equal(`"hello"`, result, "Gateway POST request not properly responding")

	status, result, err = suite.request(server, "DELETE", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(204, status, "Gateway DELETE request not properly responding")
	suite.Require().Equal(``, result, "Gateway DELETE request not properly responding")

	status, result, err = suite.request(server, "PUT", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway PUT request not properly responding")
	suite.Require().Equal(`put`, result, "Gateway PUT request not properly responding")

	status, result, err = suite.request(server, "PATCH", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway PATCH request not properly responding")
	suite.Require().Equal(`patch`, result, "Gateway PATCH request not properly responding")

	status, result, err = suite.request(server, "HEAD", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway HEAD request not properly responding")
	suite.Require().Equal(``, result, "Gateway HEAD request not properly responding")

	// Any unique path we register should have an OPTIONS that responds with 405. You can hit it
	// with middleware to do something interesting like CORS.
	status, result, err = suite.request(server, "OPTIONS", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(405, status, "Gateway missing auto-OPTIONS route")
	status, result, err = suite.request(server, "OPTIONS", "/FooService.Hello", "")
	suite.Require().NoError(err)
	suite.Require().Equal(405, status, "Gateway missing auto-OPTIONS route")

	// Should be a not-found, not a bad-method when OPTIONS is for a path we didn't register.
	status, result, err = suite.request(server, "OPTIONS", "/FooService.Goodbye", "")
	suite.Require().NoError(err)
	suite.Require().Equal(404, status, "Gateway should not accept OPTIONS for any old path/route")
}

func (suite *GatewaySuite) request(server *httptest.Server, method string, path string, body string) (int, string, error) {
	request, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}

	res, err := suite.HTTPClient.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()

	result, err := io.ReadAll(res.Body)
	return res.StatusCode, string(result), err
}

func TestGatewaySuite(t *testing.T) {
	suite.Run(t, new(GatewaySuite))
}

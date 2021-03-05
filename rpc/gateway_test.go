package rpc_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/monadicstack/frodo/rpc"
	"github.com/monadicstack/frodo/rpc/authorization"
	"github.com/monadicstack/frodo/rpc/metadata"
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
			suite.respond(w, 200, "hello")
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "get",
		Path:        "/foo/:a/:b",
		ServiceName: "FooService",
		Name:        "Sum",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			params := httptreemux.ContextParams(req.Context())
			suite.respond(w, 202, fmt.Sprintf("%s + %s = ?", params["a"], params["b"]))
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "DELETE",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Delete",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 204, "delete")
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "PUT",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Put",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 200, "put")
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "PATCH",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Patch",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 200, "patch")
		},
	})
	gateway.Register(rpc.Endpoint{
		Method:      "HEAD",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Head",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 200, "head")
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, result, err := suite.request(server, "GET", "/foo/42/43", "")
	suite.Require().NoError(err)
	suite.Require().Equal(202, status, "Gateway GET request not properly responding")
	suite.Require().Equal("42 + 43 = ?", result, "Gateway GET request not properly responding")

	status, result, err = suite.request(server, "GET", "/foo/98/99", "")
	suite.Require().NoError(err)
	suite.Require().Equal(202, status, "Gateway GET request not properly responding")
	suite.Require().Equal("98 + 99 = ?", result, "Gateway GET request not properly responding")

	status, result, err = suite.request(server, "POST", "/FooService.Hello", `{"Foo":1}`)
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway POST request not properly responding")
	suite.Require().Equal("hello", result, "Gateway POST request not properly responding")

	status, result, err = suite.request(server, "DELETE", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(204, status, "Gateway DELETE request not properly responding")
	suite.Require().Equal("", result, "Gateway DELETE request not properly responding")

	status, result, err = suite.request(server, "PUT", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway PUT request not properly responding")
	suite.Require().Equal("put", result, "Gateway PUT request not properly responding")

	status, result, err = suite.request(server, "PATCH", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway PATCH request not properly responding")
	suite.Require().Equal("patch", result, "Gateway PATCH request not properly responding")

	status, result, err = suite.request(server, "HEAD", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Gateway HEAD request not properly responding")
	suite.Require().Equal("", result, "Gateway HEAD request not properly responding")

	// Any unique path we register should have an OPTIONS that responds with 405. You can hit it
	// with middleware to do something interesting like CORS.
	status, _, err = suite.request(server, "OPTIONS", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(405, status, "Gateway missing auto-OPTIONS route")
	status, _, err = suite.request(server, "OPTIONS", "/FooService.Hello", "")
	suite.Require().NoError(err)
	suite.Require().Equal(405, status, "Gateway missing auto-OPTIONS route")

	// Should be a not-found, not a bad-method when OPTIONS is for a path we didn't register.
	status, _, err = suite.request(server, "OPTIONS", "/FooService.Goodbye", "")
	suite.Require().NoError(err)
	suite.Require().Equal(404, status, "Gateway should not accept OPTIONS for any old path/route")
}

// Ensures that you can fetch the current endpoint details from both middleware and your handler function.
func (suite *GatewaySuite) TestEndpointFromContext() {
	values := []string{"", "", ""}

	middlewareA := func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		e := rpc.EndpointFromContext(req.Context())
		values[0] = fmt.Sprintf("%s.A", e.String())
		next(w, req)
	}
	middlewareB := func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		e := rpc.EndpointFromContext(req.Context())
		values[1] = fmt.Sprintf("%s.B", e.String())
		next(w, req)
	}

	gateway := rpc.NewGateway(rpc.WithMiddleware(middlewareA, middlewareB))
	gateway.Register(rpc.Endpoint{
		Method:      "GET",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Hello",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			e := rpc.EndpointFromContext(req.Context())
			values[2] = fmt.Sprintf("%s.C", e.String())
			suite.respond(w, 200, fmt.Sprintf("ok"))
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	// Make sure that the gateway responded at all.
	status, result, err := suite.request(server, "GET", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Post-middleware handler did not properly respond")
	suite.Require().Equal("ok", result, "Post-middleware handler did not properly respond")

	// Make sure (A) middleware and the handler had access to endpoint data and (B) they did it in the right order.
	suite.Require().Equal("FooService.Hello.A", values[0], "Middleware A did not fetch endpoint properly")
	suite.Require().Equal("FooService.Hello.B", values[1], "Middleware B did not fetch endpoint properly")
	suite.Require().Equal("FooService.Hello.C", values[2], "Handler did not fetch endpoint properly")
}

// Ensures that the HTTP Authorization header passed into the handler via the Context so that your service
// function has access to it as well.
func (suite *GatewaySuite) TestAuthorizationOnContext() {
	gateway := rpc.NewGateway()
	gateway.Register(rpc.Endpoint{
		Method:      "GET",
		Path:        "/foo",
		ServiceName: "FooService",
		Name:        "Hello",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			// Just respond w/ the context's authorization info.
			auth := authorization.FromContext(req.Context())
			suite.respond(w, 200, auth.String())
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, result, err := suite.request(server, "GET", "/foo", "", func(request *http.Request) {
		request.Header.Set("Authorization", "Bearer 12345")
	})
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Authorization handler did not respond properly")
	suite.Require().Equal("Bearer 12345", result, "Authorization handler did not respond properly")

	status, result, err = suite.request(server, "GET", "/foo", "", func(request *http.Request) {
		request.Header.Set("Authorization", "DUDE!")
	})
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Authorization handler did not respond properly")
	suite.Require().Equal("DUDE!", result, "Authorization handler did not respond properly")

}

// Ensure that we respond w/ a 500 if your handler panics rather than crashing the server
func (suite *GatewaySuite) TestRecoverFromPanic_handler() {
	gateway := rpc.NewGateway()
	gateway.Register(rpc.Endpoint{
		Method: "GET",
		Path:   "/foo",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			panic("nope")
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, _, err := suite.request(server, "GET", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(500, status, "Should recover w/ 500 on handler panic")
}

// Ensure that we respond w/ a 500 if your middleware panics rather than crashing the server
func (suite *GatewaySuite) TestRecoverFromPanic_middleware() {
	gateway := rpc.NewGateway(rpc.WithMiddleware(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		panic("nope")
	}))
	gateway.Register(rpc.Endpoint{
		Method: "GET",
		Path:   "/foo",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 200, "ok")
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, _, err := suite.request(server, "GET", "/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(500, status, "Should recover w/ 500 on middleware panic")
}

// Ensure that all endpoints use the path prefix on all endpoints.
func (suite *GatewaySuite) TestGatewayPathPrefix() {
	gateway := rpc.NewGateway()
	gateway.PathPrefix = "v2"
	gateway.Register(rpc.Endpoint{
		Method: "GET",
		Path:   "/foo",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 200, "ok")
		},
	})
	gateway.Register(rpc.Endpoint{
		Method: "POST",
		Path:   "/bar",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			suite.respond(w, 200, "ok")
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, result, err := suite.request(server, "GET", "/v2/foo", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Did not respond properly when path prefix is configured")
	suite.Require().Equal("ok", result, "Did not respond properly when path prefix is configured")

	status, result, err = suite.request(server, "POST", "/v2/bar", "")
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Did not respond properly when path prefix is configured")
	suite.Require().Equal("ok", result, "Did not respond properly when path prefix is configured")

	// Make sure that the path w/o the prefix didn't sneak in there...
	status, _, err = suite.request(server, "POST", "/bar", "")
	suite.Require().NoError(err)
	suite.Require().Equal(404, status, "Did not respond properly when path prefix is configured")
}

// Ensure that metadata passed in via the X-RPC-Values header are properly added to the context.
func (suite *GatewaySuite) TestRestoreMetadata() {
	values := []string{"", ""}
	gateway := rpc.NewGateway(rpc.WithMiddleware(func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		metaString := ""
		metaBool := false
		metadata.Value(req.Context(), "metaString", &metaString)
		metadata.Value(req.Context(), "metaBool", &metaBool)
		values[0] = fmt.Sprintf("%v.%v", metaString, metaBool)

		// Curve ball - add another metadata value for the handler.
		ctx := metadata.WithValue(req.Context(), "metaInt", 42)
		next(w, req.WithContext(ctx))
	}))
	gateway.Register(rpc.Endpoint{
		Method: "GET",
		Path:   "/foo",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			metaString := ""
			metaBool := false
			metaInt := 0
			metadata.Value(req.Context(), "metaString", &metaString)
			metadata.Value(req.Context(), "metaBool", &metaBool)
			metadata.Value(req.Context(), "metaInt", &metaInt)
			values[1] = fmt.Sprintf("%v.%v.%v", metaString, metaBool, metaInt)
			suite.respond(w, 200, "ok")
		},
	})

	server := httptest.NewServer(gateway)
	defer server.Close()

	status, _, err := suite.request(server, "GET", "/foo", "", func(request *http.Request) {
		request.Header.Set(metadata.RequestHeader, `{"metaString":{"value":"A"}, "metaBool":{"value":true}}`)
	})
	suite.Require().NoError(err)
	suite.Require().Equal(200, status, "Should respond positively when accessing metadata")
	suite.Require().Equal("A.true", values[0], "Middleware did not receive proper metadata")
	suite.Require().Equal("A.true.42", values[1], "Handler did not receive proper metadata")

	// Junk data in metadata header should stop the request.
	status, _, err = suite.request(server, "GET", "/foo", "", func(request *http.Request) {
		request.Header.Set(metadata.RequestHeader, `{"met`)
	})
	suite.Require().NoError(err)
	suite.Require().Equal(400, status, "Should respond with BadRequest when metadata header is ill-formed")
}

// Ensure that EndpointFromContext returns nil when it hasn't been applied to the context yet.
func (suite *GatewaySuite) TestEndpointFromContext_missing() {
	endpoint := rpc.EndpointFromContext(nil)
	suite.Require().Nil(endpoint, "Endpoint should be nil when fetching from nil context")

	endpoint = rpc.EndpointFromContext(context.Background())
	suite.Require().Nil(endpoint, "Endpoint should be nil when not present in context")
}

func (suite *GatewaySuite) request(server *httptest.Server, method string, path string, body string, opts ...func(*http.Request)) (int, string, error) {
	request, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
	if err != nil {
		return 0, "", err
	}

	for _, opt := range opts {
		opt(request)
	}

	res, err := suite.HTTPClient.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()

	result, err := io.ReadAll(res.Body)
	return res.StatusCode, string(result), err
}

func (suite *GatewaySuite) respond(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func TestGatewaySuite(t *testing.T) {
	suite.Run(t, new(GatewaySuite))
}

package rpc

import (
	"context"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/monadicstack/frodo/rpc/authorization"
	"github.com/monadicstack/frodo/rpc/metadata"
	"github.com/monadicstack/respond"
)

// NewGateway creates a wrapper around your raw service to expose it via HTTP for RPC calls.
func NewGateway(options ...GatewayOption) Gateway {
	gw := Gateway{
		Router:     httprouter.New(),
		Binder:     jsonBinder{},
		middleware: middlewarePipeline{},
		PathPrefix: "",
		endpoints:  map[string]Endpoint{},
	}
	for _, option := range options {
		option(&gw)
	}

	// Combine all middleware (internal book-keeping and user-provided) into a single pipeline. We
	// will NOT apply them to the HandlerFunc from the router just yet. We will actually apply these
	// middlewares to every single endpoint handler inside of Register() rather than once right here.
	// We need the ROUTER to be the main entry point to your handler so that we can add the Endpoint
	// data to the request context BEFORE the other middleware fires. If we codified the middleware here
	// your service handler would have the endpoint data, but not the middleware.
	//
	// If we did gw.middleware.Then(gw.Router) here...
	//
	//   restoreEndpoint->restoreMetadata->your_middleware->ROUTER->serviceHandler
	//
	// By deferring the middleware application until we register the handler with the router we get
	// the order of operations we want:
	//
	//   ROUTER->restoreEndpoint->restoreMetadata->your_middleware->serviceHandler
	//
	// Since the router goes first, 'restoreEndpoint' has the info it needs to properly populate the context.
	mw := middlewarePipeline{
		MiddlewareFunc(recoverFromPanic),
		MiddlewareFunc(restoreEndpoint),
		MiddlewareFunc(restoreMetadata),
		MiddlewareFunc(restoreAuthorization),
	}
	gw.middleware = append(mw, gw.middleware...)
	gw.Router.SaveMatchedRoutePath = true
	return gw
}

// GatewayOption defines a setting you can apply when creating an RPC gateway via 'NewGateway'.
type GatewayOption func(*Gateway)

// Gateway wrangles all of the incoming RPC/HTTP handling for your service calls. It automatically
// converts all transport data into your Go request struct. Conversely, it also marshals and transmits
// your service response struct data back to the caller. Aside from feeding this to `http.ListenAndServe()`
// you likely won't interact with this at all yourself.
type Gateway struct {
	Name       string
	Router     *httprouter.Router
	Binder     Binder
	PathPrefix string
	middleware middlewarePipeline
	endpoints  map[string]Endpoint
}

// Register the operation with the gateway so that it can be exposed for invoking remotely.
func (gw *Gateway) Register(endpoint Endpoint) {
	// The user specified a path like "GET /user/:id" in their code, so when they fetch the
	// endpoint data later, that's what we want it to look like, so we'll leave the endpoint's
	// Path attribute alone. But... the router needs the full path which includes the optional
	// prefix (e.g. "/v2"). So we'll use the full path for routing and lookups (transparent to
	// the user), but the user will never have to see the "/v2" portion.
	path := toEndpointPath(gw.PathPrefix, endpoint.Path)

	// If you're registering "POST /FooService.Bar" we're going to create a route for
	// the POST as well as an additional, implicit OPTIONS route. This is so that
	// you can use WithMiddleware(Func) to enable CORS in your API. All of your middleware
	// is actually part of the router/mux handling (see comments in New() for details as to why), so
	// if we don't include an explicit OPTIONS route for this path then your CORS middleware
	// will never actually get invoked - httprouter will just reject the request. We fully expect
	// your CORS middleware to short-circuit the 'next' chain, so the 405 failure we're hard-coding
	// as the OPTIONS handler won't actually be invoked if you enable CORS via middleware.
	gw.endpoints[path] = endpoint
	gw.Router.HandlerFunc(strings.ToUpper(endpoint.Method), path, gw.middleware.Then(endpoint.Handler))
	gw.registerOptions(path)
}

func (gw Gateway) registerOptions(path string) {
	// I realize that recovering from panics makes the baby jesus cry. This is to handle the case where you
	// register multiple service functions with the same path, but different methods. For instance:
	//
	//   GET  /foo/bar
	//   POST /foo/bar
	//
	// Since we blindly register an options with each, we will end up registering OPTIONS twice for that
	// path. The httprouter will panic when that happens. At first I planned on just looking through the
	// gateway's already-registered endpoint paths for a match (and thus skip), but there's a case that's
	// hard to detect:
	//
	//   GET  /foo/:bar
	//   POST /foo/:goo
	//
	// A dumb string-based check would see those as unique paths, but the router will still barf because they
	// are functionally equivalent.
	//
	// So.... since Julien Schmidt is already doing all of the hard work, I'm catching the panic in this
	// instance to make life easier. If there's something fundamentally wrong with the route, we'll fail
	// more naturally when we register the "real" endpoint route, so we're not going to miss meaningful errors.
	defer func() {
		recover()
	}()
	gw.Router.HandlerFunc(http.MethodOptions, path, gw.middleware.Then(methodNotAllowedHandler{}.ServeHTTP))
}

// ServeHTTP is the central HTTP handler that includes all http routing, middleware, service forwarding, etc.
func (gw Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := context.WithValue(req.Context(), contextKeyGateway{}, &gw)
	gw.Router.ServeHTTP(w, req.WithContext(ctx))
}

// Endpoint describes an operation that we expose through an RPC gateway.
type Endpoint struct {
	// The HTTP method that should be used when exposing this endpoint in the gateway.
	Method string
	// The HTTP path pattern that should be used when exposing this endpoint in the gateway.
	Path string
	// ServiceName is the name of the service that this operation is part of.
	ServiceName string
	// Name is the name of the function/operation that this endpoint describes.
	Name string
	// Handler is the gateway function that does the "work".
	Handler http.HandlerFunc
}

// String just returns the fully qualified "Service.Operation" descriptor for the operation.
func (e Endpoint) String() string {
	return e.ServiceName + "." + e.Name
}

type contextKeyGateway struct{}
type contextKeyEndpoint struct{}

// EndpointFromContext fetches the meta information about the service RPC operation that we're currently invoking.
func EndpointFromContext(ctx context.Context) *Endpoint {
	endpoint, ok := ctx.Value(contextKeyEndpoint{}).(Endpoint)
	if !ok {
		return nil
	}
	return &endpoint
}

// recoverFromPanic automatically recovers from a panic thrown by your handler so that if you nil-pointer
// or something else unexpected, we'll safely just return a 500-style error.
func recoverFromPanic(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			respond.To(w, req).InternalServerError("%v", err)
		}
	}()
	next(w, req)
}

// restoreEndpoint places the *Endpoint data for the current operation onto the request context
// so your handler can access the RPC details about what is being invoked. Mainly useful for fetching
// logging/tracing info about the operation.
func restoreEndpoint(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	gw, ok := req.Context().Value(contextKeyGateway{}).(*Gateway)
	if !ok {
		respond.To(w, req).InternalServerError("invalid rpc gateway context")
		return
	}

	params := httprouter.ParamsFromContext(req.Context())
	endpoint, ok := gw.endpoints[params.MatchedRoutePath()]
	if !ok {
		respond.To(w, req).InternalServerError("no endpoint for path '%s'", params.MatchedRoutePath())
		return
	}

	ctx := context.WithValue(req.Context(), contextKeyEndpoint{}, endpoint)
	next(w, req.WithContext(ctx))
}

// restoreMetadata parses the "X-RPC-Values" request header and places the values onto the context's metadata
// so that all shared values from the caller are available for your handler when it's finally invoked.
func restoreMetadata(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	encodedValues := req.Header.Get(metadata.RequestHeader)

	values, err := metadata.FromJSON(encodedValues)
	if err != nil {
		respond.To(w, req).BadRequest("rpc metadata error: %v", err.Error())
		return
	}

	ctx := metadata.WithValues(req.Context(), values)
	next(w, req.WithContext(ctx))
}

// restoreAuthorization grabs the "Authorization" header from the request and puts it in the context so that
// it can be propagated across service calls. The idea is that if you call Service A with "Authorization: Bearer foo"
// and the handler ends up calling Service B, we want the underlying HTTP call to Service B to *also* have
// the same authorization header. This preserves that value so that we
func restoreAuthorization(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	auth := authorization.New(req.Header.Get("Authorization"))
	ctx := authorization.WithHeader(req.Context(), auth)
	next(w, req.WithContext(ctx))
}

// Combines the path to an endpoint (e.g. "/user/:id/contact") and an optional service
// prefix (e.g. "/v2"). The result is the complete path to this resource.
func toEndpointPath(prefix string, path string) string {
	prefix = strings.Trim(prefix, "/")
	path = strings.Trim(path, "/")

	switch prefix {
	case "":
		return "/" + path
	default:
		return "/" + prefix + "/" + path
	}
}

// methodNotAllowedHandler just replies with a 405 error status no matter what. It's the
// default OPTIONS handler we use so that you can insert the CORS middleware of your
// choice should you choose to enable browser-based communication w/ your service.
type methodNotAllowedHandler struct{}

func (methodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	respond.To(w, req).MethodNotAllowed("")
}

package rpc

import (
	"net/http"
)

/* ----- SERVER MIDDLEWARE ----- */

// WithMiddleware invokes this chain of work before executing the actual HTTP handler for
// your service call.
func WithMiddleware(mw ...Middleware) GatewayOption {
	return func(gw *Gateway) {
		gw.middleware = mw
	}
}

// WithMiddlewareFunc invokes this chain of work before executing the actual HTTP handler for
// your service call.
func WithMiddlewareFunc(funcs ...MiddlewareFunc) GatewayOption {
	mw := make(middlewarePipeline, len(funcs))
	for i, fn := range funcs {
		mw[i] = fn
	}
	return WithMiddleware(mw...)
}

// Middleware is a component that conforms to the 'negroni' middleware handler. It accepts the
// standard HTTP inputs as well as the rest of the computation.
type Middleware interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc)
}

// MiddlewareFunc is a component that conforms to the 'negroni' middleware function. It accepts the
// standard HTTP inputs as well as the rest of the computation.
type MiddlewareFunc func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc)

// ServeHTTP basically calls itself. This is a mechanism that lets middleware functions be passed
// around the same as a full middleware handler component.
func (mw MiddlewareFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	mw(w, req, next)
}

// middlewarePipeline is a chain of middleware handlers that fire in succession before ultimately
// executing the "real" HTTP handler that does the real work for the endpoint.
type middlewarePipeline []Middleware

// Then wraps all of the middleware handlers capped off with the "real work" handler into a single
// handler function that can be used by a standard net/http server.
func (pipeline middlewarePipeline) Then(handler http.HandlerFunc) http.HandlerFunc {
	for i := len(pipeline) - 1; i >= 0; i-- {
		mw := pipeline[i]
		next := handler
		handler = func(res http.ResponseWriter, req *http.Request) {
			mw.ServeHTTP(res, req, next)
		}
	}
	return handler
}

/* ----- CLIENT MIDDLEWARE ----- */

// WithClientMiddleware sets the chain of HTTP request/response handlers you want to invoke
// on each service function invocation before/after we dispatch the HTTP request.
func WithClientMiddleware(funcs ...ClientMiddleware) ClientOption {
	return func(client *Client) {
		client.middleware = funcs
	}
}

// RoundTripperFunc matches the signature of the standard http.RoundTripper interface.
type RoundTripperFunc func(r *http.Request) (*http.Response, error)

// ClientMiddleware is a round-tripper-like function that accepts a request and returns a response/error
// combo, but also accepts 'next' (the rest of the computation) so that you can short circuit the
// execution as you see fit.
type ClientMiddleware func(request *http.Request, next RoundTripperFunc) (*http.Response, error)

// clientMiddlewarePipeline is an ordered chain of client middleware handlers that should fire
// one after another.
type clientMiddlewarePipeline []ClientMiddleware

func (pipeline clientMiddlewarePipeline) Then(handler RoundTripperFunc) RoundTripperFunc {
	for i := len(pipeline) - 1; i >= 0; i-- {
		mw := pipeline[i]
		next := handler
		handler = func(request *http.Request) (*http.Response, error) {
			return mw(request, next)
		}
	}
	return handler

}

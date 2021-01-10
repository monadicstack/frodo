package rpc

import "net/http"

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

// Middleware is a component that conforms to the 'negroni' middleware function. It accepts the
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

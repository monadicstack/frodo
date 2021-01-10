package rpc

import "net/http"

type Middleware interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc)
}

type MiddlewareFunc func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc)

func (mw MiddlewareFunc) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	mw(w, req, next)
}

var nopMiddleware MiddlewareFunc = func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	next(w, req)
}

func WithMiddleware(mw Middleware) GatewayOption {
	return func(gw *Gateway) {
		gw.Middleware = mw
	}
}

func WithMiddlewareFunc(mw MiddlewareFunc) GatewayOption {
	return WithMiddleware(mw)
}

package rpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/robsignorelli/frodo/rpc/metadata"
)

// NewGateway creates a wrapper around your raw service to expose it via HTTP for RPC calls.
func NewGateway(options ...GatewayOption) Gateway {
	gw := Gateway{
		Router:     httprouter.New(),
		Binder:     jsonBinder{},
		middleware: middlewarePipeline{},
		handler:    nil,
	}
	for _, option := range options {
		option(&gw)
	}

	// Combine all middleware (internal book-keeping and user-provided) with the handler
	// for the router/mux to create a single function we'll use as the master handler when
	// we supply the gateway to ListenAndServe.
	mw := middlewarePipeline{
		MiddlewareFunc(restoreCallDetails),
		MiddlewareFunc(restoreMetadata),
	}
	mw = append(mw, gw.middleware...)
	gw.handler = mw.Then(gw.Router.ServeHTTP)

	return gw
}

// GatewayOption defines a setting you can apply when creating an RPC gateway via 'NewGateway'.
type GatewayOption func(*Gateway)

// Gateway wrangles all of the incoming RPC/HTTP handling for your service calls. It automatically
// converts all transport data into your Go request struct. Conversely, it also marshals and transmits
// your service response struct data back to the caller. Aside from feeding this to `http.ListenAndServe()`
// you likely won't interact with this at all yourself.
type Gateway struct {
	Router     *httprouter.Router
	Binder     Binder
	middleware middlewarePipeline
	handler    http.HandlerFunc
}

// ServeHTTP is the central HTTP handler that includes all http routing, middleware, service forwarding, etc.
func (gw Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.handler(w, req)
}

func restoreCallDetails(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	ctx := context.WithValue(req.Context(), "ServiceCall", "foo")
	next(w, req.WithContext(ctx))
}

func restoreMetadata(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	encodedValues := req.Header.Get(metadata.RequestHeader)

	values, err := metadata.FromJSON(encodedValues)
	if err != nil {
		http.Error(w, fmt.Sprintf("rpc metadata error: %v", err.Error()), 400)
		return
	}

	ctx := metadata.WithValues(req.Context(), values)
	next(w, req.WithContext(ctx))
}

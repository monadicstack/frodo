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
		Middleware: nopMiddleware,
	}
	for _, option := range options {
		option(&gw)
	}
	return gw
}

// Gateway wrangles all of the incoming RPC/HTTP handling for your service calls. It automatically
// converts all transport data into your Go request struct. Conversely, it also marshals and transmits
// your service response struct data back to the caller. Aside from feeding this to `http.ListenAndServe()`
// you likely won't interact with this at all yourself.
type Gateway struct {
	Router     *httprouter.Router
	Binder     Binder
	Middleware Middleware
}

func (gw Gateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// NOTE: We're actually defining these in the REVERSE order in which they'll be
	// run. Basically we want to restore all implicit RPC data before any of your
	// middleware/handler code is run.
	handler3 := gw.Router.ServeHTTP

	handler2 := func(w http.ResponseWriter, req *http.Request) {
		gw.Middleware.ServeHTTP(w, req, handler3)
	}
	handler1 := func(w http.ResponseWriter, req *http.Request) {
		restoreCallDetails(w, req, handler2)
	}
	handler := func(w http.ResponseWriter, req *http.Request) {
		restoreMetadata(w, req, handler1)
	}
	handler(w, req)
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

type GatewayOption func(*Gateway)

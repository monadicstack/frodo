package gateway

import (
	"github.com/julienschmidt/httprouter"
)

func New(options ...Option) HTTPGateway {
	gw := HTTPGateway{
		Router:     httprouter.New(),
		Binder:     JSONBinder{},
		Middleware: nopMiddleware,
	}
	for _, option := range options {
		option(&gw)
	}
	return gw
}

type HTTPGateway struct {
	Router     *httprouter.Router
	Binder     Binder
	Middleware Middleware
}

type Option func(*HTTPGateway)

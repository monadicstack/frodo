package gateway

import "github.com/julienschmidt/httprouter"

func New(options ...Option) HTTPGateway {
	gw := HTTPGateway{
		Router: httprouter.New(),
	}
	for _, option := range options {
		option(&gw)
	}
	return gw
}

type HTTPGateway struct {
	Router *httprouter.Router
}

type Option func(*HTTPGateway)

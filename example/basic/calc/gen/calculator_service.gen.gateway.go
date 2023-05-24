// Code generated by Frodo - DO NOT EDIT.
//
//   Timestamp: Tue, 10 May 2022 16:23:56 EDT
//   Source:    calc/calculator_service.go
//   Generator: https://github.com/davidrenne/frodo
//
package calc

import (
	"context"
	"net/http"

	"github.com/davidrenne/frodo/example/basic/calc"
	"github.com/davidrenne/frodo/rpc"
	"github.com/monadicstack/respond"
)

// NewCalculatorServiceGateway accepts your "real" CalculatorService instance (the thing that really does the work), and
// exposes it to other services/clients over RPC. The rpc.Gateway it returns implements http.Handler, so you
// can pass it to any standard library HTTP server of your choice.
//
//	// How to fire up your service for RPC and/or your REST API
//	service := calc.CalculatorService{ /* set up to your liking */ }
//	gateway := calc.NewCalculatorServiceGateway(service)
//	http.ListenAndServe(":8080", gateway)
//
// The default instance works well enough, but you can supply additional options such as WithMiddleware() which
// accepts any negroni-compatible middleware handlers.
func NewCalculatorServiceGateway(service calc.CalculatorService, options ...rpc.GatewayOption) CalculatorServiceGateway {
	gw := rpc.NewGateway(options...)
	gw.Name = "CalculatorService"
	gw.PathPrefix = ""

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/CalculatorService.Add",
		ServiceName: "CalculatorService",
		Name:        "Add",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := calc.AddRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.Add(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/CalculatorService.Sub",
		ServiceName: "CalculatorService",
		Name:        "Sub",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := calc.SubRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.Sub(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	return CalculatorServiceGateway{Gateway: gw, service: service}
}

type CalculatorServiceGateway struct {
	rpc.Gateway
	service calc.CalculatorService
}

func (gw CalculatorServiceGateway) Add(ctx context.Context, request *calc.AddRequest) (*calc.AddResponse, error) {
	return gw.service.Add(ctx, request)
}

func (gw CalculatorServiceGateway) Sub(ctx context.Context, request *calc.SubRequest) (*calc.SubResponse, error) {
	return gw.service.Sub(ctx, request)
}

func (gw CalculatorServiceGateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.Gateway.ServeHTTP(w, req)
}

// Code generated by Frodo from name_service.go - DO NOT EDIT
//
//   https://github.com/monadicstack/frodo
//
package names

import (
	"context"
	"net/http"

	"github.com/monadicstack/frodo/example/names"
	"github.com/monadicstack/frodo/rpc"
	"github.com/monadicstack/respond"
)

// NewNameServiceGateway accepts your "real" NameService instance (the thing that really does the work), and
// exposes it to other services/clients over RPC. The rpc.Gateway it returns implements http.Handler, so you
// can pass it to any standard library HTTP server of your choice.
//
//	// How to fire up your service for RPC and/or your REST API
//	service := names.NameService{ /* set up to your liking */ }
//	gateway := names.NewNameServiceGateway(service)
//	http.ListenAndServe(":8080", gateway)
//
// The default instance works well enough, but you can supply additional options such as WithMiddleware() which
// accepts any negroni-compatible middleware handlers.
func NewNameServiceGateway(service names.NameService, options ...rpc.GatewayOption) NameServiceGateway {
	gw := rpc.NewGateway(options...)
	gw.Name = "NameService"
	gw.PathPrefix = ""

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/NameService.Download",
		ServiceName: "NameService",
		Name:        "Download",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := names.DownloadRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.Download(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/NameService.DownloadExt",
		ServiceName: "NameService",
		Name:        "DownloadExt",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := names.DownloadExtRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.DownloadExt(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/NameService.FirstName",
		ServiceName: "NameService",
		Name:        "FirstName",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := names.FirstNameRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.FirstName(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/NameService.LastName",
		ServiceName: "NameService",
		Name:        "LastName",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := names.LastNameRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.LastName(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/NameService.SortName",
		ServiceName: "NameService",
		Name:        "SortName",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := names.SortNameRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.SortName(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/NameService.Split",
		ServiceName: "NameService",
		Name:        "Split",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := names.SplitRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.Split(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	return NameServiceGateway{gateway: gw, service: service}
}

type NameServiceGateway struct {
	gateway rpc.Gateway
	service names.NameService
}

func (gw NameServiceGateway) Download(ctx context.Context, request *names.DownloadRequest) (*names.DownloadResponse, error) {
	return gw.service.Download(ctx, request)
}

func (gw NameServiceGateway) DownloadExt(ctx context.Context, request *names.DownloadExtRequest) (*names.DownloadExtResponse, error) {
	return gw.service.DownloadExt(ctx, request)
}

func (gw NameServiceGateway) FirstName(ctx context.Context, request *names.FirstNameRequest) (*names.FirstNameResponse, error) {
	return gw.service.FirstName(ctx, request)
}

func (gw NameServiceGateway) LastName(ctx context.Context, request *names.LastNameRequest) (*names.LastNameResponse, error) {
	return gw.service.LastName(ctx, request)
}

func (gw NameServiceGateway) SortName(ctx context.Context, request *names.SortNameRequest) (*names.SortNameResponse, error) {
	return gw.service.SortName(ctx, request)
}

func (gw NameServiceGateway) Split(ctx context.Context, request *names.SplitRequest) (*names.SplitResponse, error) {
	return gw.service.Split(ctx, request)
}

func (gw NameServiceGateway) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	gw.gateway.ServeHTTP(w, req)
}

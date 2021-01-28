// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from post_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package postsrpc

import (
	"net/http"

	"github.com/robsignorelli/frodo/example/posts"
	"github.com/robsignorelli/frodo/rpc"
	"github.com/robsignorelli/respond"
)

// NewPostServiceGateway accepts your "real" PostService instance (the thing that really does the work), and
// exposes it to other services/clients over RPC. The rpc.Gateway it returns implements http.Handler, so you
// can pass it to any standard library HTTP server of your choice.
//
//	// How to fire up your service for RPC and/or your REST API
//	service := posts.PostService{ /* set up to your liking */ }
//	gateway := postsrpc.NewPostServiceGateway(service)
//	http.ListenAndServe(":8080", gateway)
//
// The default instance works well enough, but you can supply additional options such as WithMiddleware() which
// accepts any negroni-compatible middleware handlers.
func NewPostServiceGateway(service posts.PostService, options ...rpc.GatewayOption) rpc.Gateway {
	gw := rpc.NewGateway(options...)
	gw.Name = "PostService"
	gw.PathPrefix = "v2"

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/PostService.GetPost",
		ServiceName: "PostService",
		Name:        "GetPost",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := posts.GetPostRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.GetPost(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/PostService.CreatePost",
		ServiceName: "PostService",
		Name:        "CreatePost",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := posts.CreatePostRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.CreatePost(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/PostService.Archive",
		ServiceName: "PostService",
		Name:        "Archive",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := posts.ArchiveRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.Archive(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	return gw
}

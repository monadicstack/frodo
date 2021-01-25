// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated server code from comment_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package commentsrpc

import (
	"net/http"

	"github.com/robsignorelli/frodo/example/comments"
	"github.com/robsignorelli/frodo/rpc"
	"github.com/robsignorelli/respond"
)

// NewCommentServiceGateway accepts your "real" CommentService instance (the thing that really does the work), and
// exposes it to other services/clients over RPC. The rpc.Gateway it returns implements http.Handler, so you
// can pass it to any standard library HTTP server of your choice.
//
//	// How to fire up your service for RPC and/or your REST API
//	service := comments.CommentService{ /* set up to your liking */ }
//	gateway := commentsrpc.NewCommentServiceGateway(service)
//	http.ListenAndServe(":8080", gateway)
//
// The default instance works well enough, but you can supply additional options such as WithMiddleware() which
// accepts any negroni-compatible middleware handlers.
func NewCommentServiceGateway(service comments.CommentService, options ...rpc.GatewayOption) rpc.Gateway {
	gw := rpc.NewGateway(options...)
	gw.Name = "CommentService"

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/CommentService.CreateComment",
		ServiceName: "CommentService",
		Name:        "CreateComment",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := comments.CreateCommentRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.CreateComment(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	gw.Register(rpc.Endpoint{
		Method:      "POST",
		Path:        "/CommentService.FindByPost",
		ServiceName: "CommentService",
		Name:        "FindByPost",
		Handler: func(w http.ResponseWriter, req *http.Request) {
			response := respond.To(w, req)

			serviceRequest := comments.FindByPostRequest{}
			if err := gw.Binder.Bind(req, &serviceRequest); err != nil {
				response.Fail(err)
				return
			}

			serviceResponse, err := service.FindByPost(req.Context(), &serviceRequest)
			response.Reply(200, serviceResponse, err)
		},
	})

	return gw
}

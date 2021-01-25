// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated client code from comment_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package commentsrpc

import (
	"context"
	"fmt"

	"github.com/robsignorelli/frodo/example/comments"
	"github.com/robsignorelli/frodo/rpc"
)

// NewCommentServiceClient creates an RPC client that conforms to the CommentService interface, but delegates
// work to remote instances. You must supply the base address of the remote service gateway instance or
// the load balancer for that service.
//
// You should be able to get a working service using default options (i.e. no options), but you can customize
// the HTTP client, define middleware, and more using client options. All of the ones that apply to the RPC
// client are named "WithClientXXX()".
func NewCommentServiceClient(address string, options ...rpc.ClientOption) *CommentServiceClient {
	return &CommentServiceClient{
		Client: rpc.NewClient("CommentService", address, options...),
	}
}

// CommentServiceClient manages all interaction w/ a remote CommentService instance by letting you invoke functions
// on this instance as if you were doing it locally (hence... RPC client). You shouldn't instantiate this
// manually. Instead, you should utilize the NewCommentServiceClient() function to properly set this up.
type CommentServiceClient struct {
	rpc.Client
}

func (client *CommentServiceClient) CreateComment(ctx context.Context, request *comments.CreateCommentRequest) (*comments.CreateCommentResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &comments.CreateCommentResponse{}
	err := client.Invoke(ctx, "POST", "/CommentService.CreateComment", request, response)
	return response, err
}

func (client *CommentServiceClient) FindByPost(ctx context.Context, request *comments.FindByPostRequest) (*comments.FindByPostResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &comments.FindByPostResponse{}
	err := client.Invoke(ctx, "POST", "/CommentService.FindByPost", request, response)
	return response, err
}

// CommentServiceProxy fully implements the CommentService interface, but delegates all operations to a "real"
// instance of the service. You can embed this type in a struct of your choice so you can "override" or
// decorate operations as you see fit. Any operations on CommentService that you don't explicitly define will
// simply delegate to the default implementation of the underlying 'Service' value.
//
// Since you have access to the underlying service, you are able to both implement custom handling logic AND
// call the "real" implementation, so this can be used as special middleware that applies to only certain operations.
type CommentServiceProxy struct {
	Service comments.CommentService
}

func (proxy *CommentServiceProxy) CreateComment(ctx context.Context, request *comments.CreateCommentRequest) (*comments.CreateCommentResponse, error) {
	return proxy.Service.CreateComment(ctx, request)
}

func (proxy *CommentServiceProxy) FindByPost(ctx context.Context, request *comments.FindByPostRequest) (*comments.FindByPostResponse, error) {
	return proxy.Service.FindByPost(ctx, request)
}

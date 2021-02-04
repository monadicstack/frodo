// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated client code from example/posts/post_service.go
// !!!!!!! DO NOT EDIT !!!!!!!
package postsrpc

import (
	"context"
	"fmt"

	"github.com/robsignorelli/frodo/example/posts"
	"github.com/robsignorelli/frodo/rpc"
)

// NewPostServiceClient creates an RPC client that conforms to the PostService interface, but delegates
// work to remote instances. You must supply the base address of the remote service gateway instance or
// the load balancer for that service.
// PostService is a service that manages blog/article posts. This is just for example purposes,
// so this is not a truly exhaustive set of operations that you might want if you were *really*
// building some sort of blog/CRM engine.
func NewPostServiceClient(address string, options ...rpc.ClientOption) *PostServiceClient {
	rpcClient := rpc.NewClient("PostService", address, options...)
	rpcClient.PathPrefix = "v2"
	return &PostServiceClient{Client: rpcClient}
}

// PostServiceClient manages all interaction w/ a remote PostService instance by letting you invoke functions
// on this instance as if you were doing it locally (hence... RPC client). You shouldn't instantiate this
// manually. Instead, you should utilize the NewPostServiceClient() function to properly set this up.
type PostServiceClient struct {
	rpc.Client
}

// GetPost fetches a Post record by its unique identifier.
func (client *PostServiceClient) GetPost(ctx context.Context, request *posts.GetPostRequest) (*posts.GetPostResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &posts.GetPostResponse{}
	err := client.Invoke(ctx, "GET", "post/:id", request, response)
	return response, err
}

// CreatePost creates/stores a new Post record.
func (client *PostServiceClient) CreatePost(ctx context.Context, request *posts.CreatePostRequest) (*posts.CreatePostResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &posts.CreatePostResponse{}
	err := client.Invoke(ctx, "POST", "post", request, response)
	return response, err
}

// Archive effectively disables a Post from appearing in the article list.
func (client *PostServiceClient) Archive(ctx context.Context, request *posts.ArchiveRequest) (*posts.ArchiveResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("precondition failed: nil context")
	}
	if request == nil {
		return nil, fmt.Errorf("precondition failed: nil request")
	}

	response := &posts.ArchiveResponse{}
	err := client.Invoke(ctx, "PATCH", "post/:id/archive", request, response)
	return response, err
}

// PostServiceProxy fully implements the PostService interface, but delegates all operations to a "real"
// instance of the service. You can embed this type in a struct of your choice so you can "override" or
// decorate operations as you see fit. Any operations on PostService that you don't explicitly define will
// simply delegate to the default implementation of the underlying 'Service' value.
//
// Since you have access to the underlying service, you are able to both implement custom handling logic AND
// call the "real" implementation, so this can be used as special middleware that applies to only certain operations.
type PostServiceProxy struct {
	Service posts.PostService
}

func (proxy *PostServiceProxy) GetPost(ctx context.Context, request *posts.GetPostRequest) (*posts.GetPostResponse, error) {
	return proxy.Service.GetPost(ctx, request)
}

func (proxy *PostServiceProxy) CreatePost(ctx context.Context, request *posts.CreatePostRequest) (*posts.CreatePostResponse, error) {
	return proxy.Service.CreatePost(ctx, request)
}

func (proxy *PostServiceProxy) Archive(ctx context.Context, request *posts.ArchiveRequest) (*posts.ArchiveResponse, error) {
	return proxy.Service.Archive(ctx, request)
}

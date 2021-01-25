package main

import (
	"net/http"

	"github.com/robsignorelli/frodo/example/comments"
	"github.com/robsignorelli/frodo/example/comments/gen"
	postsrpc "github.com/robsignorelli/frodo/example/posts/gen"
)

func main() {
	// You just use standard dependency injection to feed the post service RPC client
	// to the comment service handler. It's just a "PostService" implementation.
	commentsService := comments.CommentServiceHandler{
		PostService: postsrpc.NewPostServiceClient("http://localhost:9001"),
	}
	gateway := commentsrpc.NewCommentServiceGateway(&commentsService)
	http.ListenAndServe(":9002", gateway)
}

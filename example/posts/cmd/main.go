package main

import (
	"net/http"

	"github.com/robsignorelli/frodo/example/posts"
	"github.com/robsignorelli/frodo/example/posts/gen"
)

func main() {
	postService := posts.PostServiceHandler{}
	gateway := postsrpc.NewPostServiceGateway(&postService)
	http.ListenAndServe(":9001", gateway)
}

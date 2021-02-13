package main

import (
	"fmt"
	"net/http"

	"github.com/monadicstack/frodo/example/posts"
	postsrpc "github.com/monadicstack/frodo/example/posts/gen"
	"github.com/monadicstack/frodo/rpc"
)

func main() {
	postService := posts.PostServiceHandler{}
	gateway := postsrpc.NewPostServiceGateway(&postService, rpc.WithMiddleware(Echo, Echo2, Echo3))
	http.ListenAndServe(":9001", gateway)
}

func Echo(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("Handling: 1A", req.Method, req.URL)
	next(w, req)
	fmt.Println("Handling: 1B", req.Method, req.URL)
}

func Echo2(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("Handling: 2A", req.Method, req.URL)
	next(w, req)
	fmt.Println("Handling: 2B", req.Method, req.URL)
}

func Echo3(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("Handling: 3A", req.Method, req.URL)
	next(w, req)
	fmt.Println("Handling: 3B", req.Method, req.URL)
}

package main

import (
	"fmt"
	"net/http"

	"github.com/robsignorelli/frodo/example/posts"
	"github.com/robsignorelli/frodo/example/posts/gen"
)

func main() {
	postService := posts.PostServiceHandler{}
	gateway := postsrpc.NewPostServiceGateway(&postService)
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

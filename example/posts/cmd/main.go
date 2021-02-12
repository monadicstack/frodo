package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robsignorelli/frodo/example/posts"
	"github.com/robsignorelli/frodo/example/posts/gen"
)

func main() {
	//postService := posts.PostServiceHandler{}

	postService := postsrpc.MockPostService{
		GetPostFunc: func(ctx context.Context, request *posts.GetPostRequest) (*posts.GetPostResponse, error) {
			if request.ID == "donnie" {
				return nil, fmt.Errorf("you're out of your element")
			}
			return &posts.GetPostResponse{ID: request.ID}, nil
		},
		CreatePostFunc: nil,
		ArchiveFunc:    nil,
	}

	gateway := postsrpc.NewPostServiceGateway(&postService)

	go func() {
		time.Sleep(15 * time.Second)
		fmt.Println(">>>>>> Times Get", postService.Calls.GetPost.Times())
		fmt.Println(">>>>>> Times Get:1", postService.Calls.GetPost.TimesFor(posts.GetPostRequest{ID: "1"}))
		fmt.Println(">>>>>> Times Get:2", postService.Calls.GetPost.TimesFor(posts.GetPostRequest{ID: "2"}))
		fmt.Println(">>>>>> Times Get:donnie", postService.Calls.GetPost.TimesFor(posts.GetPostRequest{ID: "donnie"}))
		fmt.Println(">>>>>> Times Get:len=1", postService.Calls.GetPost.TimesMatching(func(request posts.GetPostRequest) bool {
			return len(request.ID) == 1
		}))
		fmt.Println(">>>>>> Times Get:len=2", postService.Calls.GetPost.TimesMatching(func(request posts.GetPostRequest) bool {
			return len(request.ID) == 2
		}))
		fmt.Println(">>>>>> Times Create", postService.Calls.CreatePost.Times())
		fmt.Println(">>>>>> Times Create:none", postService.Calls.CreatePost.TimesFor(posts.CreatePostRequest{Title: "asldjfk"}))
	}()

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

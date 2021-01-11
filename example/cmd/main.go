package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robsignorelli/frodo/example"
	"github.com/robsignorelli/frodo/example/gen"
	"github.com/robsignorelli/frodo/rpc"
	"github.com/robsignorelli/frodo/rpc/metadata"
)

func main() {
	groupService := example.GroupServiceServer{}
	gw := examplerpc.NewGroupServiceGateway(groupService,
		rpc.WithMiddlewareFunc(
			Logger,
			Logger2,
		),
		rpc.WithPrefix("v2"),
	)

	go runClientTest()
	http.ListenAndServe(":8080", gw)
}

func runClientTest() {
	ctx := context.Background()
	ctx = metadata.WithValue(ctx, "foo", 12345)
	ctx = metadata.WithValue(ctx, "bar", example.GetByIDRequest{
		ID:   "abcdef",
		Flag: true,
	})

	time.Sleep(1 * time.Second)
	client := examplerpc.NewGroupServiceClient("http://localhost:8080",
		rpc.WithClientMiddleware(
			ClientLogger,
			ClientLogger2,
		),
		rpc.WithClientPathPrefix("v2"),
	)
	response, err := client.GetByID(ctx, &example.GetByIDRequest{
		ID:   "123x45",
		Flag: false,
	})
	fmt.Printf(">>>>>> %+v : %v", response, err)
}

func Logger(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("> Hello", req.URL.String())
	next(w, req)
	fmt.Println("> Goodbye")
}

func Logger2(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("> Hello 2")
	next(w, req)
	fmt.Println("> Goodbye 2")
}

func ClientLogger(request *http.Request, next rpc.RoundTripperFunc) (*http.Response, error) {
	fmt.Println(">>>>> BEFORE CLIENT INVOKE")
	r, err := next(request)
	fmt.Println(">>>>> AFTER CLIENT INVOKE")
	return r, err
}

func ClientLogger2(request *http.Request, next rpc.RoundTripperFunc) (*http.Response, error) {
	fmt.Println(">>>>> BEFORE CLIENT INVOKE 2")
	r, err := next(request)
	fmt.Println(">>>>> AFTER CLIENT INVOKE 2")
	return r, err
}

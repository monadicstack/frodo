package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robsignorelli/frodo/example"
	"github.com/robsignorelli/frodo/rpc"
	"github.com/robsignorelli/frodo/rpc/metadata"
)

func main() {
	groupService := example.GroupServiceServer{}
	gw := example.NewGroupServiceGateway(groupService,
		rpc.WithMiddlewareFunc(
			Logger,
			Logger2,
		),
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

	time.Sleep(2 * time.Second)
	client := example.NewGroupServiceClient("http://localhost:8080")
	response, err := client.GetByID(ctx, &example.GetByIDRequest{
		ID:   "123x45",
		Flag: false,
	})
	fmt.Printf(">>>>>> %+v : %v", response, err)
}

func Logger(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("> Hello")
	next(w, req)
	fmt.Println("> Goodbye")
}

func Logger2(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("> Hello 2")
	next(w, req)
	fmt.Println("> Goodbye 2")
}

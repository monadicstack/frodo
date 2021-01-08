package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robsignorelli/expose/example"
	"github.com/robsignorelli/expose/gateway"
)

func main() {
	groupService := example.GroupServiceServer{}
	gw := example.NewGroupServiceGateway(groupService,
		gateway.WithMiddlewareFunc(Logger),
	)

	go runClientTest()
	http.ListenAndServe(":8080", gw)
}

func runClientTest() {
	time.Sleep(2 * time.Second)
	client := example.NewGroupServiceClient("http://localhost:8080")
	response, err := client.GetByID(context.Background(), &example.GetByIDRequest{
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

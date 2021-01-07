package main

import (
	"fmt"
	"net/http"

	"github.com/robsignorelli/expose/example"
	"github.com/robsignorelli/expose/gateway"
)

func main() {
	groupService := example.GroupServiceServer{}
	gw := example.NewGroupServiceGateway(groupService,
		gateway.WithMiddlewareFunc(Logger),
	)

	http.ListenAndServe(":8080", gw)
}

func Logger(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Println("> Hello")
	next(w, req)
	fmt.Println("> Goodbye")
}

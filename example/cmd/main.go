package main

import (
	"net/http"

	"github.com/robsignorelli/expose/example"
)

func main() {
	groupService := example.GroupServiceServer{}
	gw := example.NewGroupServiceGateway(groupService)

	http.ListenAndServe(":8080", gw)
}

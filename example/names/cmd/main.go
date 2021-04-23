package main

import (
	"net/http"

	"github.com/monadicstack/frodo/example/names"
	namesrpc "github.com/monadicstack/frodo/example/names/gen"
)

func main() {
	serviceHandler := names.NameServiceHandler{}
	gateway := namesrpc.NewNameServiceGateway(&serviceHandler)
	http.ListenAndServe(":9100", gateway)
}

package main

import (
	"net/http"

	"github.com/davidrenne/frodo/example/basic/calc"
	calcrpc "github.com/davidrenne/frodo/example/basic/calc/gen"
)

func main() {
	serviceHandler := calc.CalculatorServiceHandler{}
	gateway := calcrpc.NewCalculatorServiceGateway(&serviceHandler)
	http.ListenAndServe(":9000", gateway)
}

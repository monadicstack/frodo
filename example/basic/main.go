package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davidrenne/frodo/example/basic/calc"
	calcrpc "github.com/davidrenne/frodo/example/basic/calc/gen"
)

func main() {
	ctx := context.Background()
	client := calcrpc.NewCalculatorServiceClient("http://localhost:9000")

	add, err := client.Add(ctx, &calc.AddRequest{A: 5, B: 2})
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("5 + 2 = ", add.Result)

	sub, err := client.Sub(ctx, &calc.SubRequest{A: 5, B: 2})
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("5 - 2 = ", sub.Result)
}

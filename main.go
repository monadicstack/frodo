package main

import (
	"fmt"
	"log"
	"os"

	"github.com/robsignorelli/frodo/generate"
	"github.com/robsignorelli/frodo/parser"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatalf("Usage: frodoc [go-file]")
	}

	inputFileName := os.Args[1]

	log.Printf("[frodoc] Parsing service definitions: %s", inputFileName)
	ctx, err := parser.ParseFile(inputFileName)
	crapPants(err)

	log.Printf("[frodoc] Generating artifacts...")
	crapPants(generate.Artifact(ctx, inputFileName, generate.TemplateGatewayGo))
	crapPants(generate.Artifact(ctx, inputFileName, generate.TemplateClientGo))
}

func crapPants(err error) {
	if err != nil {
		fmt.Printf("[frodoc] fatal error: %v\n", err)
		os.Exit(1)
	}
}

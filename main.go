package main

import (
	"fmt"
	"os"

	"github.com/robsignorelli/frodo/cli"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "frodo",
		Short: "A code generator for Go-based (micro)services that creates RPC clients/gateways.",
	}
	rootCmd.AddCommand(cli.GenerateGateway{}.Command())
	rootCmd.AddCommand(cli.GenerateClient{}.Command())
	rootCmd.AddCommand(cli.GenerateMock{}.Command())
	rootCmd.AddCommand(cli.CreateService{}.Command())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/rego"
	"github.com/walter-manger/rego_va/builtins"
)

func main() {
	rego.RegisterBuiltin1(builtins.RegisterIdentity())
	rego.RegisterBuiltin1(builtins.RegisterResource())
	rego.RegisterBuiltin1(builtins.RegisterCheck())

	if err := cmd.RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

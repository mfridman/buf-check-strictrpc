package main

import (
	"github.com/bufbuild/bufplugin-go/check"
	"github.com/mfridman/buf-lint-strictrpc/internal/strictrpc"
)

func main() {
	rules := []*check.RuleSpec{
		strictrpc.Rule,
	}
	check.Main(&check.Spec{Rules: rules})
}

package main

import (
	"github.com/bufbuild/bufplugin-go/check"
	"github.com/mfridman/buf-lint-strictrpc/internal/strictrpc"
)

func main() {
	check.Main(strictrpc.Spec)
}

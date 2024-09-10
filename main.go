package main

import (
	"buf.build/go/bufplugin/check"
	"github.com/mfridman/buf-lint-strictrpc/internal/strictrpc"
)

func main() {
	check.Main(strictrpc.Spec)
}

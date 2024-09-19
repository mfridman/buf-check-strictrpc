# buf-check-strictrpc [![Actions](https://github.com/mfridman/buf-check-strictrpc/workflows/CI/badge.svg)](https://github.com/mfridman/buf-check-strictrpc/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/mfridman/buf-check-strictrpc)](https://goreportcard.com/report/github.com/mfridman/buf-check-strictrpc)

This repository contains a custom `buf` check plugin designed to enforce a stricter set of lint
rules for Protobuf RPCs, particularly when using frameworks like gRPC or
[Connect](https://connectrpc.com).

While the `buf` linter already enforces some of these rules, this plugin introduces additional
constraints:

1. Files containing a service definition must have a filename ending with `_service.proto`.
1. Each file can only contain one `service` definition, which must appear at the top before any
   message definitions.
1. For every RPC method, there must be exactly two associated messages, named after the method and
   suffixed with `Request` and `Response`.
1. Optionally, a third message may be present, which must be suffixed with `ErrorDetails`.
1. Messages in the file must be listed in the same order as their corresponding RPC methods.

### Options

Useful plugin options to configure the plugin:

| Option                 | Description                                                       | Default |
| ---------------------- | ----------------------------------------------------------------- | ------- |
| `disable_streaming`    | Disables streaming RPCs. (Some people just don't like em)         | `false` |
| `allow_protobuf_empty` | Allows usage of `google.protobuf.Empty` in requests or responses. | `false` |

### Example

Given a `dragon_service.proto` file:

```proto
service DragonService {
  rpc GetDragon(GetDragonRequest) returns (GetDragonResponse) {}
  rpc TrainDragon(TrainDragonRequest) returns (TrainDragonResponse) {}
}

message GetDragonRequest {}

message GetDragonResponse {}

message TrainDragonRequest {}

message TrainDragonResponse {}

// This message is optional.
message TrainDragonErrorDetails {}
```

### Preparing a WASM binary

Compile the Go plugin to a WASM binary:

```bash
GOOS=wasip1 GOARCH=wasm go build -o buf-check-strictrpc.wasm main.go
```

Let's try it out!

```bash
wasmtime buf-check-strictrpc.wasm --help

Usage of plugin:
      --format string   The format to use for requests, responses, and specs. Must be one of ["binary", "json"]. (default "binary")
      --protocol        Print the protocol to stdout and exit.
      --spec            Print the spec to stdout in the specified format and exit.
```

Neat, let's see what the protocol looks like:

```bash
wasmtime buf-check-strictrpc.wasm --protocol
1
```

Okay, now let's try something more interesting, what does the spec look like?

```bash
wasmtime buf-check-strictrpc.wasm --spec --format=json | jq
```

Output:

```json
{
  "procedures": [
    {
      "path": "/buf.plugin.check.v1.CheckService/Check",
      "args": ["check"]
    },
    {
      "path": "/buf.plugin.check.v1.CheckService/ListRules",
      "args": ["list-rules"]
    },
    {
      "path": "/buf.plugin.check.v1.CheckService/ListCategories",
      "args": ["list-categories"]
    }
  ]
}
```

### Using WASM binary with `buf`

We could compile the plugin to normal Go binary, but what's the fun in that? So let's use wasmtime
to run the WASM binary:

```bash
GOOS=wasip1 GOARCH=wasm go build -o ./examples/buf-check-strictrpc.wasm main.go

buf lint ./examples/dragon.proto
examples/dragon.proto:1:1:service "dragon" must end with _service.proto. (wasmtime ./examples/buf-check-strictrpc.wasm)
```

Yikes, we forgot to add the `_service.proto` suffix to our file!

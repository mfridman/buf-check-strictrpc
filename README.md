# buf-check-strictrpc

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

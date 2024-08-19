# buf-lint-strictrpc

This repository contains a custom buf lint plugin for enforcing a specific set of lint rules when
writing Protobuf RPCs, using frameworks like gRPC or [Connect](https://connectrpc.com/).

Some of this is already enforced by the buf linter, but we take it a step further:

1. A file with a Service definition must end with `_service.proto`.
2. Within this file, there must be exactly one `service` definition, defined at the top.
3. For every RPC method, there must be exactly two messages, starting with the method name and
   suffixed with `Request` and `Response`.
4. Optionally, a third message may exist and must be suffixed with `ErrorDetails`.
5. The order of messages is based on the method definitions.

### Example

Given a `user_service.proto` file:

```proto
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {}
}

message GetUserRequest {}

message GetUserResponse {}

// This message is optional.
message GetUserErrorDetails {}

message CreateUserRequest {}

message CreateUserResponse {}

// This message is optional.
message CreateUserErrorDetails {}
```

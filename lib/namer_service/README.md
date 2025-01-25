# Notes

Remember:

``` shell
go get google.golang.org/grpc
```

To build from the base dir (check in the proto file itself for latest build instructions)

``` shell
protoc --go_out=. --go-grpc_out=. lib/namer_service/namer_service.proto
```

# GO-Delivery Simple Microservices

## Install Protoc

```bash
go get -u -v google.golang.org/protobuf/...
go get -u -v google.golang.org/grpc/...

go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

export PATH="$PATH:$(go env GOPATH)/bin"
```

### Generate ProtoBuffers

```bash
protoc --go_out=. --go-grpc_out=. ./messages/*.proto
```
#!/bin/bash

echo "building protobuffers..."
protoc --go_out=. --go-grpc_out=. ./messages/*.proto
echo "protobuffers built."

echo "cleaning cache..."
go clean --cache
echo "cache cleaned."

echo "running tests..."
go test --cover go-delivery/util/...
go test --cover go-delivery/db/...
go test --cover go-delivery/security/...
go test --cover go-delivery/services/...
echo "tests finished."

echo "cleanup bin..."
rm bin/*

echo "building services..."
go build -o bin/accounts services/accounts/main.go
go build -o bin/wallets  services/wallets/main.go
go build -o bin/sellers  services/sellers/main.go
go build -o bin/orders   services/orders/main.go
go build -o bin/api      services/api/main.go
echo "services built."

chmod +x bin/*
tree bin/
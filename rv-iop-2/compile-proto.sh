#!/bin/bash

# Generate Go code from rv-iop proto files
# This script should be run from the project root directory

# Ensure protoc and plugins are installed:
# - protoc: https://github.com/protocolbuffers/protobuf/releases
# - protoc-gen-go: go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.2
# - protoc-gen-go-grpc: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

cd client/proto || exit 1

protoc \
  --proto_path=. \
  --go_out=../grpc/rv-iop \
  --go_opt=paths=source_relative \
  --go-grpc_out=../grpc/rv-iop \
  --go-grpc_opt=paths=source_relative \
  rv-iop/*.proto


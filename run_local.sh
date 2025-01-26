#!/bin/bash
mkdir -p ./bin || true
go build -o ./bin/vsl_secrets ./cmd/main.go
./bin/vsl_secrets
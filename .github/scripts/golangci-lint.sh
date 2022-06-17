#!/bin/bash

cd tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
results=$( golangci-lint run ./internal/... 2>&1 )
echo "${results}"

cd providerlint && results=$( golangci-lint run ./... 2>&1 )
echo "${results}"

#!/usr/bin/env bash

set -euo pipefail

go_version="$(goenv version)"
echo "Go version: ${go_version}"
echo "go env:"
go env

echo "compiling test binary"
go test -c -o ./test-binary

# If there are no test files in a package, `go test -c` will succeed, but no executable will be generated
if [[ ! -f ./test-binary ]]; then
    echo "##teamcity[buildStop comment='No tests in this package']"
fi

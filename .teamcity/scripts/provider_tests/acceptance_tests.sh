#!/usr/bin/env bash

set -euo pipefail

# All of internal except for internal/service. This list should be generated.
TF_ACC=1 go test \
    ./internal/acctest/... \
    ./internal/conns/... \
    ./internal/create/... \
    ./internal/experimental/... \
    ./internal/flex/... \
    ./internal/generate/... \
    ./internal/provider/... \
    ./internal/tags/... \
    ./internal/tfresource/... \
    ./internal/vault/... \
    ./internal/verify/... \
    -json -v -count=1 -parallel "%ACCTEST_PARALLELISM%" -timeout=0 -run=TestAcc

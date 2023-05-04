#!/usr/bin/env bash

set -euo pipefail

# shellcheck disable=2157 # This isn't a constant string, it's a TeamCity variable substitution
if [[ -n "%ACCTEST_ROLE_ARN%" ]]; then
    conf=$(pwd)/aws.conf

    function cleanup {
        rm "${conf}"
    }
    trap cleanup EXIT

    touch "${conf}"
    chmod 600 "${conf}"
    cat <<EOF >"${conf}"
[profile primary]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = primary_user

[profile primary_user]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=primary
fi

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

#!/usr/bin/env bash

set -euo pipefail

# shellcheck disable=2157 # This isn't a constant string, it's a TeamCity variable substitution
if [[ -n "%ACCTEST_ROLE_ARN%" ]]; then
    echo "assuming role %ACCTEST_ROLE_ARN% for sweeper"

    echo "AWS_ACCESS_KEY_ID: $(echo "%AWS_ACCESS_KEY_ID%" | sed -E "s/^.+(.{4})/****\1/")"
    echo "AWS_SECRET_ACCESS_KEY: $(echo "%AWS_SECRET_ACCESS_KEY%" | sed -E "s/^(.{4}).+(.{4})/\1****\2/")"
    echo "ACCTEST_ROLE_ARN: %ACCTEST_ROLE_ARN%"

    conf=$(pwd)/aws.conf

    function cleanup {
        rm "${conf}"
    }
    trap cleanup EXIT

    touch "${conf}"
    chmod 600 "${conf}"
    cat <<EOF >"${conf}"
[profile sweeper]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = source

[profile source]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=sweeper
fi

echo "AWS_CONFIG_FILE: ${AWS_CONFIG_FILE}"
echo "AWS_PROFILE: ${AWS_PROFILE}"
echo "env AWS_ACCESS_KEY_ID: $(echo "${AWS_ACCESS_KEY_ID}" | sed -E "s/^(.{4}).+(.{4})/\1****\2/")"

go test ./internal/sweep -v -tags=sweep -sweep="%SWEEPER_REGIONS%" -sweep-allow-failures -timeout=4h

#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

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
[profile perftest]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = source

[profile source]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY

    export AWS_CONFIG_FILE="${conf}"
    export AWS_PROFILE=perftest
fi

export PACKAGE_PARALLELISM=$(( "%teamcity.agent.hardware.cpuCount%" / 2 ))

exec "%teamcity.build.checkoutDir%/.ci/scripts/smoke-tests-identity.sh"

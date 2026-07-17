#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [[ "%PKG%" == "" ]]; then
	echo "PKG variable is required"
	exit 1
fi

PKG="./internal/service/%PKG%/..."

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [[ "%TEST_PATTERN%" == "" || "%TEST_PATTERN%" == "TestAcc" ]]; then
	echo "Invalid test filter pattern: \"%TEST_PATTERN%\""
	exit 1
fi

# shellcheck disable=2157 # These aren't constant strings, they're TeamCity variable substitution
if [[ -n "%ACCTEST_ROLE_ARN%" || -n "%ACCTEST_ALTERNATE_ROLE_ARN%" ]]; then
	conf=$(pwd)/aws.conf

	function cleanup {
		rm "${conf}"
	}
	trap cleanup EXIT

	touch "${conf}"
	chmod 600 "${conf}"

	export AWS_CONFIG_FILE="${conf}"

	# shellcheck disable=2157 # This isn't a constant string, it's a TeamCity variable substitution
	if [[ -n "%ACCTEST_ROLE_ARN%" ]]; then
		cat <<EOF >>"${conf}"
[profile primary]
role_arn       = %ACCTEST_ROLE_ARN%
source_profile = primary_user

[profile primary_user]
aws_access_key_id     = %AWS_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_SECRET_ACCESS_KEY%
EOF

		unset AWS_ACCESS_KEY_ID
		unset AWS_SECRET_ACCESS_KEY

		export AWS_PROFILE=primary
	fi

	# shellcheck disable=2157 # This isn't a constant string, it's a TeamCity variable substitution
	if [[ -n "%ACCTEST_ALTERNATE_ROLE_ARN%" ]]; then
		cat <<EOF >>"${conf}"
[profile alternate]
role_arn       = %ACCTEST_ALTERNATE_ROLE_ARN%
source_profile = alternate_user

[profile alternate_user]
aws_access_key_id     = %AWS_ALTERNATE_ACCESS_KEY_ID%
aws_secret_access_key = %AWS_ALTERNATE_SECRET_ACCESS_KEY%
EOF

		export AWS_ALTERNATE_PROFILE=alternate
	fi
fi

echo "Downloading Go module dependencies"
go mod download

echo "Running acceptance tests for ${PKG} with pattern %TEST_PREFIX%"

TEST_PREFIX="%TEST_PREFIX%"
TEST_PREFIX="${TEST_PREFIX#\(}"
TEST_PREFIX="${TEST_PREFIX%\)}"
echo "make testacc PKG=%PKG% TESTARGS='-run=${TEST_PREFIX}'" > /tmp/test_command.txt

TF_ACC=1 go test "${PKG}" -count=1 -json -v -run="%TEST_PREFIX%" -parallel "%ACCTEST_PARALLELISM%" -timeout=0 -vet=off -buildvcs=false \
    | tee /tmp/test_output.json

jq -s '[.[] | select(.Action == "pass" or .Action == "fail" or .Action == "skip") | select(.Test != null)] | length' \
    /tmp/test_output.json > /tmp/test_count.txt

echo "Total tests run: $(cat /tmp/test_count.txt)"

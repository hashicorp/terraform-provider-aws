#!/usr/bin/env bash

set -euo pipefail

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [[ "%PKG%" != "" ]]; then
	TEST_PKG="./internal/service/%PKG%/..."
elif [[ "%PKG_NAME%" != "" ]]; then
	TEST_PKG="./%PKG_NAME%/..."
else
	echo "One of PKG, for a service package, or PKG_NAME, for other packages, must be specified."
	exit 1
fi

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [[ "%TEST_PREFIX%" == "" || "%TEST_PREFIX%" == "TestAcc" ]]; then
	echo "Invalid test filter pattern: \"%TEST_PREFIX%\""
	exit 1
fi

echo "Filtering acceptance tests: %TEST_PREFIX%"

TEST_LIST=$(go test "${TEST_PKG}" -list="%TEST_PREFIX%" 2>/dev/null)

read -r -a split <<<"${TEST_LIST}"
TEST_COUNT=${#split[@]}

if [[ "${TEST_COUNT}" == 0 ]]; then
	echo "Zero tests"
	exit 0
elif [[ "${TEST_COUNT}" == 1 ]]; then
	echo "Running 1 test:"
else
	echo "Running ${TEST_COUNT} tests:"
fi
echo "${TEST_LIST}"
echo

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

TF_ACC=1 go test "${TEST_PKG}" -run="%TEST_PREFIX%" -json -count=1 -parallel "%ACCTEST_PARALLELISM%" -timeout=0

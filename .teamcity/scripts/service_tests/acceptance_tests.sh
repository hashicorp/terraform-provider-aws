#!/usr/bin/env bash

set -euo pipefail

TEST_LIST=$(./test-binary -test.list="%TEST_PATTERN%" 2>/dev/null)

if [[ -n "%TEST_EXCLUDE_PATTERN%" ]]; then
  TEST_LIST=$(echo "${TEST_LIST}" | grep -vE "%TEST_EXCLUDE_PATTERN%")
fi

read -r -a split <<<"${TEST_LIST}"
TEST_COUNT=${#split[@]}

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [[ "%TEST_PATTERN%" != "TestAcc" ]]; then
	echo "Filtering acceptance tests: %TEST_PATTERN%"
fi
if [[ "${TEST_COUNT}" == 0 ]]; then
	echo "Zero acceptance tests"
	exit 0
elif [[ "${TEST_COUNT}" == 1 ]]; then
	echo "Running 1 acceptance test:"
else
	echo "Running ${TEST_COUNT} acceptance tests:"
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

		unset AWS_ALTERNATE_ACCESS_KEY_ID
		unset AWS_ALTERNATE_SECRET_ACCESS_KEY

		export AWS_ALTERNATE_PROFILE=alternate
	fi
fi

echo "${TEST_LIST}" | TF_ACC=1 teamcity-go-test -test ./test-binary -parallelism "%ACCTEST_PARALLELISM%"

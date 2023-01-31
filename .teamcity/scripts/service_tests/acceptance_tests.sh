#!/usr/bin/env bash

set -euo pipefail

TEST_LIST=$(./test-binary -test.list="%TEST_PATTERN%" 2>/dev/null)

read -r -a split <<<"${TEST_LIST}"
TEST_COUNT=${#split[@]}

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [ "%TEST_PATTERN%" != "TestAcc" ]; then
	echo "Filtering acceptance tests: %TEST_PATTERN%"
fi
if [ "$TEST_COUNT" == 0 ]; then
	echo "Zero acceptance tests"
	exit 0
elif [ "$TEST_COUNT" == 1 ]; then
	echo "Running 1 acceptance test:"
else
	echo "Running ${TEST_COUNT} acceptance tests:"
fi
echo "${TEST_LIST}"
echo

echo "${TEST_LIST}" | TF_ACC=1 teamcity-go-test -test ./test-binary -parallelism "%ACCTEST_PARALLELISM%"

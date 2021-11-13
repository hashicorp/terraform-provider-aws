#!/usr/bin/env bash

set -euo pipefail

if [[ "%TEST_PATTERN%" == "" || "%TEST_PATTERN%" == "TestAcc" ]]; then
  echo "Invalid test filter pattern: \"%TEST_PATTERN%\""
  exit 1
fi

echo "Filtering acceptance tests: %TEST_PATTERN%"

TEST_LIST=$(go test ./... -list="%TEST_PATTERN%" 2>/dev/null)

read -r -a split <<<"${TEST_LIST}"
TEST_COUNT=${#split[@]}

if [ "$TEST_COUNT" == 0 ]; then
	echo "Zero tests"
	exit 0
elif [ "$TEST_COUNT" == 1 ]; then
	echo "Running 1 test:"
else
	echo "Running ${TEST_COUNT} tests:"
fi
echo "${TEST_LIST}"
echo

TF_ACC=1 go test ./... -run="%TEST_PATTERN%" -v -count=1 -parallelism "%ACCTEST_PARALLELISM%" -timeout=0

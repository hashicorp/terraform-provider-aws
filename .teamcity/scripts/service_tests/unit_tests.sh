#!/usr/bin/env bash
# shellcheck disable=SC2162

set -euo pipefail

# Go doesn't support negative lookahead regexes, so we have to fake it
TEST_LIST=$(./test-binary -test.list='Test([^A]|A[^c]|Ac[^c])' 2>/dev/null)

read -a split <<<"${TEST_LIST}"
TEST_COUNT=${#split[@]}

if [[ "${TEST_COUNT}" == 0 ]]; then
	echo "Zero unit tests"
	exit 0
fi

echo "${TEST_LIST}" | teamcity-go-test -test ./test-binary

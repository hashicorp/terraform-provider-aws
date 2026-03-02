#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -uo pipefail

# Check test naming conventions for generated and list tests

check_tests() {
	local test_type="$1"
	local description="$2"

	local total uppercase lowercase correct

	total=$(rg '^func [tT]estAcc[A-Z]' -t go internal/ 2>/dev/null | grep "${test_type}:" | wc -l | xargs)
	
	if [ "$total" -eq 0 ]; then
		echo "Error: No ${description} found. Check ripgrep command."
		return 1
	fi

	uppercase=$(rg '^func [tT]estAcc[A-Z][^_]*(_[A-Za-z][^_]*)*_[A-Z][^_(]*\(' -t go internal/ 2>/dev/null | grep "${test_type}:" | wc -l | xargs)
	lowercase=$(rg '^func [tT]estAcc[A-Z][^_]*(_[a-zA-Z][^_]*)*(_[a-z][^_]*)+(_[a-zA-Z][^_]*)*_[a-zA-Z][^_(]*\(' -t go internal/ 2>/dev/null | grep "${test_type}:" | wc -l | xargs)
	correct=$((total - uppercase - lowercase))

	echo "${description}: ${correct}/${total} correct"

	local has_errors=0

	if [ "$uppercase" -gt 0 ]; then
		echo "Error: Found ${uppercase} tests with uppercase final segment:"
		rg '^func [tT]estAcc[A-Z][^_]*(_[A-Za-z][^_]*)*_[A-Z][^_(]*\(' -t go internal/ 2>/dev/null | grep "${test_type}:" || true
		has_errors=1
	fi

	if [ "$lowercase" -gt 0 ]; then
		echo "Error: Found ${lowercase} tests with lowercase middle segment:"
		rg '^func [tT]estAcc[A-Z][^_]*(_[a-zA-Z][^_]*)*(_[a-z][^_]*)+(_[a-zA-Z][^_]*)*_[a-zA-Z][^_(]*\(' -t go internal/ 2>/dev/null | grep "${test_type}:" || true
		has_errors=1
	fi

	return $has_errors
}

command -v rg >/dev/null 2>&1 || { echo "Error: ripgrep (rg) is required but not installed."; exit 1; }

echo "Checking test naming conventions..."

exit_code=0

check_tests "_gen_test.go" "Generated tests (*_gen_test.go)" || exit_code=1
check_tests "_list_test.go" "List tests (*_list_test.go)" || exit_code=1

if [ $exit_code -eq 0 ]; then
	echo "✓ All test names follow correct convention"
fi

exit $exit_code

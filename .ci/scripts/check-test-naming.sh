#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -uo pipefail

# Check test naming conventions for generated and list tests

readonly all_pattern='^func [tT]estAcc[A-Z]'
readonly uppercase_pattern='^func [tT]estAcc[A-Z][^_]*(_[A-Za-z][^_]*)*_[A-Z][^_(]*\('
readonly lowercase_pattern='^func [tT]estAcc[A-Z][^_]*(_[a-zA-Z][^_]*)*(_[a-z][^_]*)+(_[a-zA-Z][^_]*)*_[a-zA-Z][^_(]*\('

# search prints matching lines in tracked and untracked (but not ignored) files.
# git grep exits 1 when nothing matches, which is not an error here.
search() {
	local pattern="$1"
	local test_type="$2"

	git grep --untracked -E "${pattern}" -- "internal/*${test_type}"
	local rc=$?
	if [ "$rc" -gt 1 ]; then
		echo "Error: git grep failed (exit ${rc})" >&2
		return 2
	fi
}

count() {
	search "$@" | wc -l | xargs
}

check_tests() {
	local test_type="$1"
	local description="$2"

	local total uppercase lowercase correct

	total=$(count "${all_pattern}" "${test_type}") || return 1

	if [ "$total" -eq 0 ]; then
		echo "Error: No ${description} found. Check git grep command."
		return 1
	fi

	uppercase=$(count "${uppercase_pattern}" "${test_type}") || return 1
	lowercase=$(count "${lowercase_pattern}" "${test_type}") || return 1
	correct=$((total - uppercase - lowercase))

	echo "${description}: ${correct}/${total} correct"

	local has_errors=0

	if [ "$uppercase" -gt 0 ]; then
		echo "Error: Found ${uppercase} tests with uppercase final segment:"
		search "${uppercase_pattern}" "${test_type}"
		has_errors=1
	fi

	if [ "$lowercase" -gt 0 ]; then
		echo "Error: Found ${lowercase} tests with lowercase middle segment:"
		search "${lowercase_pattern}" "${test_type}"
		has_errors=1
	fi

	return $has_errors
}

command -v git >/dev/null 2>&1 || { echo "Error: git is required but not installed."; exit 1; }

cd "$(git rev-parse --show-toplevel)" || exit 1

echo "Checking test naming conventions..."

exit_code=0

check_tests "_gen_test.go" "Generated tests (*_gen_test.go)" || exit_code=1
check_tests "_list_test.go" "List tests (*_list_test.go)" || exit_code=1

if [ $exit_code -eq 0 ]; then
	echo "✓ All test names follow correct convention"
fi

exit $exit_code

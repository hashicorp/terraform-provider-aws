#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Tests the live-test runner with stub executables; no AWS API calls are made.
set -euo pipefail
umask 077

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
test_dir="$(mktemp -d)"
trap 'rm -rf "$test_dir"' EXIT
mkdir "$test_dir/bin"

cat > "$test_dir/bin/go" <<'EOF'
#!/usr/bin/env bash
if [[ "$1" == version ]]; then
  printf 'go version go1.26.0 linux/amd64\n'
else
  printf '%s\n' "$TEST_OUTPUT"
fi
EOF
cat > "$test_dir/bin/terraform" <<'EOF'
#!/usr/bin/env bash
printf 'Terraform v1.14.0\n'
EOF
cat > "$test_dir/bin/uname" <<'EOF'
#!/usr/bin/env bash
printf 'Linux\n'
EOF
chmod +x "$test_dir/bin/"*

test_name=TestAccODBDBNodeAction_existingNode
pass="--- PASS: $test_name (0.01s)"
skip="--- SKIP: $test_name (0.01s)"
cases=0

check_result() {
  local label="$1" expected="$2" output="$3" actual=0
  env -i PATH="$test_dir/bin:/usr/bin:/bin" TMPDIR="$test_dir" \
    AWS_PROFILE=runner-test GO_BINARY="$test_dir/bin/go" \
    TF_ACC_TERRAFORM_PATH="$test_dir/bin/terraform" TEST_OUTPUT="$output" \
    bash "$script_dir/test-odb-db-node-actions.sh" us-east-1 cluster-test node-test \
    > "$test_dir/output" 2>&1 || actual=$?
  if (( actual != expected )); then
    printf '%s: expected exit %s, got %s\n' "$label" "$expected" "$actual" >&2
    cat "$test_dir/output" >&2
    exit 1
  fi
  if (( expected == 0 )); then
    grep -q 'Live test passed' "$test_dir/output"
  elif grep -q 'Live test passed' "$test_dir/output"; then
    printf '%s: a skipped or incomplete test was reported as passed\n' "$label" >&2
    exit 1
  fi
  cases=$((cases + 1))
}

check_result pass 0 "$pass"
check_result child_pass 0 "$pass
    --- PASS: $test_name/actions (0.01s)"
check_result root_skip 1 "$skip"
check_result child_skip 1 "$pass
    --- SKIP: $test_name/actions (0.01s)"
check_result nested_skip 1 "$pass
    --- PASS: $test_name/actions (0.01s)
        --- SKIP: $test_name/actions/nested (0.01s)"
check_result unrelated_skip 0 "$pass
--- SKIP: ${test_name}_other (0.01s)"
check_result wrong_test 1 "--- PASS: ${test_name}_other (0.01s)"
check_result child_only 1 "    --- PASS: $test_name/actions (0.01s)"
printf '%s runner regression cases passed.\n' "$cases"

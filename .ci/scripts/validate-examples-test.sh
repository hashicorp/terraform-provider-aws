#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
test_dir=$(mktemp -d "${TMPDIR:-/tmp}/validate-examples-test.XXXXXX")
trap 'rm -rf "$test_dir"' EXIT

examples_dir="$test_dir/examples"
mkdir -p "$examples_dir/odb/autonomous-database" "$examples_dir/other example"
touch "$examples_dir/odb/main.tf" "$examples_dir/odb/autonomous-database/main.tf" "$examples_dir/other example/main.tf"
mkdir -p "$examples_dir/odb/.terraform/modules/downloaded"
touch "$examples_dir/odb/.terraform/modules/downloaded/main.tf"
printf 'example = "value"\n' > "$examples_dir/odb/terraform.template.tfvars"

export TERRAFORM_BIN="$test_dir/terraform"
export VALIDATION_LOG="$test_dir/commands.log"
export TEST_TERRAFORM_VERSION
export FAIL_COMMAND=""
cat > "$TERRAFORM_BIN" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "$1" == version ]]; then
    printf 'Terraform v%s\n' "$TEST_TERRAFORM_VERSION"
    exit 0
fi
printf '%s|%s\n' "${PWD##*/}" "$*" >> "$VALIDATION_LOG"
if [[ "$1" == "$FAIL_COMMAND" ]]; then
    exit 17
fi
EOF
chmod +x "$TERRAFORM_BIN"

for version in 0.12.31 1.0.6 1.11.4; do
    TEST_TERRAFORM_VERSION="$version"
    : > "$VALIDATION_LOG"
    bash "$script_dir/validate-examples.sh" "$version" "$examples_dir" > "$test_dir/output.log"
    if [[ "$version" == 1.11.4 ]]; then
        directories=(autonomous-database)
    else
        directories=(odb "other example")
    fi
    : > "$test_dir/expected.log"
    for directory in "${directories[@]}"; do
        printf '%s|init -backend=false -input=false\n' "$directory" >> "$test_dir/expected.log"
        printf '%s|fmt -check\n' "$directory" >> "$test_dir/expected.log"
        printf '%s|validate\n' "$directory" >> "$test_dir/expected.log"
    done
    diff -u "$test_dir/expected.log" "$VALIDATION_LOG"
done
cmp "$examples_dir/odb/terraform.template.tfvars" "$examples_dir/odb/terraform.tfvars"

# Failures must reach CI unchanged, without validating later phases or examples.
TEST_TERRAFORM_VERSION=1.0.6
for command in init fmt validate; do
    FAIL_COMMAND="$command"
    : > "$VALIDATION_LOG"
    status=0
    bash "$script_dir/validate-examples.sh" 1.0.6 "$examples_dir" > "$test_dir/output.log" 2>&1 || status=$?
    if [[ "$status" -ne 17 ]]; then
        printf 'Expected %s failure status 17, got %s\n' "$command" "$status" >&2
        exit 1
    fi
    case "$command" in
        init) expected_count=1 ;;
        fmt) expected_count=2 ;;
        validate) expected_count=3 ;;
    esac
    actual_count=$(wc -l < "$VALIDATION_LOG")
    if [[ "$actual_count" -ne "$expected_count" ]]; then
        printf 'Validation continued after %s failed\n' "$command" >&2
        exit 1
    fi
done
FAIL_COMMAND=""

# A mismatched executable must not silently validate with the wrong version.
: > "$VALIDATION_LOG"
if bash "$script_dir/validate-examples.sh" 1.11.4 "$examples_dir" > "$test_dir/output.log" 2>&1; then
    echo 'Expected a Terraform version mismatch to fail' >&2
    exit 1
fi
test ! -s "$VALIDATION_LOG"

echo 'Examples validation script tests passed.'

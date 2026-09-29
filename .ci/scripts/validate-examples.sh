#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -euo pipefail

# Keep the supported versions in sync with .github/workflows/examples.yml.
terraform_version=${1:?Usage: validate-examples.sh <terraform-version> [examples-directory]}
examples_dir=${2:-./examples}
examples_dir=${examples_dir%/}
terraform_bin=${TERRAFORM_BIN:-terraform}
modern_dir="$examples_dir/odb/autonomous-database"

case "$terraform_version" in
    0.12.31|1.0.6) modern=false ;;
    1.11.4) modern=true ;;
    *)
        printf 'Unsupported examples validation version: %s\n' "$terraform_version" >&2
        exit 1
        ;;
esac

actual_version=$("$terraform_bin" version | sed -n '1s/^Terraform v//p')
if [[ "$actual_version" != "$terraform_version" ]]; then
    printf 'Expected Terraform %s, got %s\n' "$terraform_version" "$actual_version" >&2
    exit 1
fi

# Ignore downloaded modules when rerunning locally. Preserve spaces in names and
# propagate discovery failures.
directories=$(find "$examples_dir" -type d -name .terraform -prune -o -type f -name '*.tf' -exec dirname {} \; | sort -u)
count=0
while IFS= read -r directory; do
    [[ -n "$directory" ]] || continue
    case "$directory" in
        "$modern_dir"|"$modern_dir/"*) [[ "$modern" == true ]] || continue ;;
        *) [[ "$modern" == false ]] || continue ;;
    esac
    (
        cd "$directory"
        if [[ -f terraform.template.tfvars ]]; then
            cp terraform.template.tfvars terraform.tfvars
        fi
        printf '\n===> Initializing Example: %s (Terraform %s) <===\n' "$directory" "$terraform_version"
        "$terraform_bin" init -backend=false -input=false
        printf '\n===> Format Checking Example: %s <===\n' "$directory"
        "$terraform_bin" fmt -check
        printf '\n===> Validating Example: %s <===\n' "$directory"
        "$terraform_bin" validate
    )
    count=$((count + 1))
done <<< "$directories"

if [[ "$count" -eq 0 ]]; then
    echo 'No examples selected for validation.' >&2
    exit 1
fi
printf '\nValidated %s example directories with Terraform %s.\n' "$count" "$terraform_version"

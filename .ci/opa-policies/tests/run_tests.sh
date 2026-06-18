#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Integration tests for OPA policies.
# Each test case is a directory containing a main.tf file.
# Directories named "pass" or "no_resource" are expected to produce no issues.
# All other directories are expected to produce issues (non-zero exit).

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
POLICY_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
TFLINT_CONFIG="${SCRIPT_DIR}/tflint.hcl"

exit_code=0

for test_dir in "${SCRIPT_DIR}"/*/; do
    for case_dir in "${test_dir}"*/; do
        case_name=$(basename "${case_dir}")
        rule_name=$(basename "${test_dir}")

        # Determine expected result from directory name
        case "${case_name}" in
            pass*|no_resource*)
                expect_pass=true
                ;;
            *)
                expect_pass=false
                ;;
        esac

        # Run tflint
        set +e
        output=$(tflint --config "${TFLINT_CONFIG}" --chdir="${case_dir}" 2>&1)
        result=$?
        set -e

        if [[ "${expect_pass}" == "true" ]] && [[ ${result} -ne 0 ]]; then
            echo "FAIL: ${rule_name}/${case_name} (expected pass, got failure)"
            echo "${output}"
            echo
            exit_code=1
        elif [[ "${expect_pass}" == "false" ]] && [[ ${result} -eq 0 ]]; then
            echo "FAIL: ${rule_name}/${case_name} (expected failure, got pass)"
            echo
            exit_code=1
        else
            echo "OK:   ${rule_name}/${case_name}"
        fi
    done
done

exit "${exit_code}"

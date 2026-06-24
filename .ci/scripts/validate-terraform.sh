#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

set -eo pipefail

trap 'exit 130' INT TERM

# Validate embedded Terraform/HCL blocks inside Go test files (or any other
# files terrafmt understands) with tflint.
#
# Two modes:
#
#   1. Default - read filenames from stdin and validate each in parallel.
#
#        find ./internal -name '*_test.go' | ./.ci/scripts/validate-terraform.sh
#
#      Set the JOBS env var to control parallelism (default: number of cores).
#
#   2. Internal - validate a single file. Mode 1 fans out to mode 2 to spread
#      the work across cores via xargs -P.
#
#        ./.ci/scripts/validate-terraform.sh --process-file <path>

if [[ -z "${TERRAFMT_CMD}" ]]; then TERRAFMT_CMD="terrafmt"; fi
if [[ -z "${TFLINT_CMD}" ]]; then TFLINT_CMD="tflint"; fi

# tflint resolves config paths relative to the working directory.
TFLINT_CONFIG="${TFLINT_CONFIG:-$(pwd -P)/.ci/.tflint.hcl}"
TFLINT_OPA_CONFIG="${TFLINT_OPA_CONFIG:-$(pwd -P)/.ci/.tflint_opa.hcl}"
export TERRAFMT_CMD TFLINT_CMD TFLINT_CONFIG TFLINT_OPA_CONFIG

# Rules used for the standard (non-OPA) pass. --only is incompatible with the
# OPA plugin, so OPA rules run in a separate tflint invocation.
RULES=(
    # Syntax checks
    "--only=terraform_comment_syntax"
    "--only=terraform_deprecated_index"
    "--only=terraform_deprecated_interpolation"
    "--only=terraform_deprecated_lookup"
    "--only=terraform_empty_list_equality"
    # Ensure valid instance types
    "--only=aws_db_instance_invalid_type"
    # Ensure modern instance types
    "--only=aws_db_instance_previous_type"
    "--only=aws_instance_previous_type"
    # Ensure engine types are valid
    "--only=aws_db_instance_invalid_engine"
    "--only=aws_mq_broker_invalid_engine_type"
)

process_one_file() {
    local filename=$1
    local exit_code=0
    local block_number=0

    while IFS= read -r block ; do
        ((block_number+=1))

        # One jq call extracts both line numbers; a second writes the block
        # text straight to the temp file. (Down from three jq forks.)
        local meta start_line end_line td tf
        meta=$(jq -r '"\(.start_line)\t\(.end_line)"' <<< "${block}")
        IFS=$'\t' read -r start_line end_line <<< "${meta}"

        td=$(mktemp -d)
        tf="${td}/main.tf"
        jq -r '.text' <<< "${block}" > "${tf}"

        local tflint_output tflint_exitcode

        # Standard rules pass.
        set +e
        tflint_output=$(${TFLINT_CMD} --config "${TFLINT_CONFIG}" --chdir="${td}" "${RULES[@]}" 2>&1)
        tflint_exitcode=$?
        set -e
        if [[ ${tflint_exitcode} -ne 0 ]]; then
            printf 'ERROR: File "%s", block #%d (lines %s-%s):\n%s\n\n' \
                "${filename}" "${block_number}" "${start_line}" "${end_line}" "${tflint_output}"
            exit_code=1
        fi

        # OPA rules pass.
        set +e
        tflint_output=$(${TFLINT_CMD} --config "${TFLINT_OPA_CONFIG}" --chdir="${td}" 2>&1)
        tflint_exitcode=$?
        set -e
        if [[ ${tflint_exitcode} -ne 0 ]] && [[ "${tflint_output}" != *"eval_builtin_error"* ]]; then
            printf 'ERROR: File "%s", block #%d (lines %s-%s):\n%s\n\n' \
                "${filename}" "${block_number}" "${start_line}" "${end_line}" "${tflint_output}"
            exit_code=1
        fi

        rm -rf "${td}"
    done < <( ${TERRAFMT_CMD} blocks --fmtcompat --json "${filename}" | jq --compact-output '.blocks[]?' )

    return "${exit_code}"
}

# Single-file mode.
if [[ "${1:-}" == "--process-file" ]]; then
    process_one_file "$2"
    exit $?
fi

# Default (orchestrator) mode: read filenames from stdin and fan out.
JOBS="${JOBS:-$(nproc 2>/dev/null || echo 4)}"

# xargs -P exits 123 if any worker exits non-zero (1-125), so the workflow
# step still fails when any file has lint errors.
xargs -P "${JOBS}" -n 1 -I {} "$0" --process-file "{}"

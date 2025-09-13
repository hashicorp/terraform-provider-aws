#!/usr/bin/env bash

set -eo pipefail

# This script works from stdin and expects one filename per line.
# To call it, e.g.
# find ./website/docs -type f \( -name '*.md' -o -name '*.markdown' \) \
#   | ./.ci/scripts/validate-terraform.sh

if [[ -z "${TERRAFMT_CMD}" ]]; then TERRAFMT_CMD="terrafmt"; fi
if [[ -z "${TFLINT_CMD}" ]]; then TFLINT_CMD="tflint"; fi

exit_code=0

# tflint always resolves config flies relative to the working directory when using --recursive
TFLINT_CONFIG="$(pwd -P)/.ci/.tflint.hcl"

# Configure the rules for tflint.
rules=(
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
while read -r filename ; do
    block_number=0

    while IFS= read -r block ; do
        ((block_number+=1))
        start_line=$(echo "${block}" | jq '.start_line')
        end_line=$(echo "${block}" | jq '.end_line')
        text=$(echo "${block}" | jq --raw-output '.text')

        td=$(mktemp -d)
        tf="${td}/main.tf"

        echo "${text}" > "${tf}"

        # We need to capture the output and error code here. We don't want to exit on the first error
        set +e
        tflint_output=$(${TFLINT_CMD} --config "${TFLINT_CONFIG}" --chdir="${td}" "${rules[@]}" 2>&1)
        tflint_exitcode=$?
        set -e

        if [[ ${tflint_exitcode} -ne 0 ]]; then
            echo "ERROR: File \"${filename}\", block #${block_number} (lines ${start_line}-${end_line}):"
            echo "${tflint_output}"
            echo
            exit_code=1
        fi
    done < <( ${TERRAFMT_CMD} blocks --fmtcompat --json "${filename}" | jq --compact-output '.blocks[]?' )
done

exit "${exit_code}"

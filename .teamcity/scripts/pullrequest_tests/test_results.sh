#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# shellcheck disable=2050 # This isn't a constant string, it's a TeamCity variable substitution
if [[ "%POST_GITHUB_COMMENT%" != "true" ]]; then
    exit 0
fi

fetch_results() {
    curl -s -u "%system.teamcity.auth.userId%:%system.teamcity.auth.password%" \
        -H "Accept: application/json" \
        "%teamcity.serverUrl%/app/rest/testOccurrences?locator=build:(id:%teamcity.build.id%),count:100000&fields=testOccurrence(name,status,duration,details)"
}

# Poll until the TeamCity index contains all tests recorded by the test runner
expected_count=$(cat /tmp/test_count.txt)
while true; do
    response=$(fetch_results)
    curr_count=$(echo "${response}" | jq '.testOccurrence | length')
    if [[ "${curr_count}" -eq "${expected_count}" ]]; then
        break
    fi
    sleep 5
done

# Authenticate internally using TeamCity's system-provided tokens
results=$(echo "${response}" |
	jq -r '.testOccurrence[] | "\(.name | sub(".*(?<t>TestAcc.*)"; "\(.t)")): [\(.status | if . == "SUCCESS" then "PASS" elif . == "FAILURE" then "FAIL" else . end)] \(.duration/1000)s\(if .status == "FAILURE" and .details != null then "\n\(.details)" else "" end)"')

echo "${results}"

# "%BRANCH_NAME%" is in the format "refs/pull/48516/merge"
pr_number="$(echo "%BRANCH_NAME%" | sed -E 's#refs/pull/([0-9]+)/merge#\1#')"
if [[ ! "${pr_number}" =~ ^[0-9]+$ ]]; then
    echo "Could not determine PR number from branch: %BRANCH_NAME%"
    exit 1
fi

gh="%system.teamcity.build.checkoutDir%/tools/gh"

go_cmd="$(cat /tmp/test_command.txt)"
body="$(printf '### Latest automated test results:\n\n```console\n%s\n\n%s\n```' "${go_cmd}" "${results}")"

"${gh}" pr comment "${pr_number}" --body "${body}"

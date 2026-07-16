#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

fetch_results() {
    curl -s -u "%system.teamcity.auth.userId%:%system.teamcity.auth.password%" \
        -H "Accept: application/json" \
        "%teamcity.serverUrl%/app/rest/testOccurrences?locator=build:(id:%teamcity.build.id%),count:100000&fields=testOccurrence(name,status,duration)"
}

# Poll until the result count stabilises across two consecutive checks
prev_count=-1
while true; do
    response=$(fetch_results)
    curr_count=$(echo "${response}" | jq '.testOccurrence | length')
    if [[ "${curr_count}" -eq "${prev_count}" ]]; then
        break
    fi
    prev_count="${curr_count}"
    sleep 5
done

# Authenticate internally using TeamCity's system-provided tokens
results=$(echo "${response}" |
	jq -r '.testOccurrence[] | "\(.name | sub(".*(?<t>TestAcc.*)"; "\(.t)")): [\(.status)] \(.duration/1000)s"')

echo "${results}"

"%BRANCH_NAME%" is in the format "refs/pull/48516/merge"
pr_number="$(echo "%BRANCH_NAME%" | sed -E 's#refs/pull/([0-9]+)/merge#\1#')"

body="$(printf '```console\n%s\n```' "${results}")"

"%system.teamcity.build.checkoutDir%/tools/gh" pr comment "${pr_number}" --body "${body}"

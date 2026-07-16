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

# "%BRANCH_NAME%" is in the format "refs/pull/48516/merge"
pr_number="$(echo "%BRANCH_NAME%" | sed -E 's#refs/pull/([0-9]+)/merge#\1#')"

gh="%system.teamcity.build.checkoutDir%/tools/gh"
marker="<!-- tc-test-results -->"
repo="$("${gh}" repo view --json nameWithOwner --jq '.nameWithOwner')"

body="$(printf '%s\n```console\n%s\n```' "${marker}" "${results}")"

# Update existing comment if one was previously posted, otherwise create a new one
comment_id="$("${gh}" api "repos/${repo}/issues/${pr_number}/comments" \
    --jq ".[] | select(.body | contains(\"${marker}\")) | .id")"

if [[ -n "${comment_id}" ]]; then
    "${gh}" api "repos/${repo}/issues/comments/${comment_id}" \
        --method PATCH \
        --field body="${body}"
else
    "${gh}" pr comment "${pr_number}" --body "${body}"
fi

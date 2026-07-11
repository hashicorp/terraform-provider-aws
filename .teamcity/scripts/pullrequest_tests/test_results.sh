#!/usr/bin/env bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

sleep 10
# Authenticate internally using TeamCity's system-provided tokens
results=$(curl -s -u "%system.teamcity.auth.userId%:%system.teamcity.auth.password%" \
     -H "Accept: application/json" \
     "%teamcity.serverUrl%/app/rest/testOccurrences?locator=build:(id:%teamcity.build.id%),count:100000&fields=testOccurrence(name,status,duration)" |
	jq -r '.testOccurrence[] | "\(.name | sub(".*(?<t>TestAcc.*)"; "\(.t)")): [\(.status)] \(.duration/1000)s"')

echo "${results}"

# "%BRANCH_NAME%" is in the format "refs/pull/48516/merge"
#pr_number="$(echo "%BRANCH_NAME%" | sed -E 's#refs/pull/([0-9]+)/merge#\1#')"
#
#body="$(printf '```console\n%s\n```' "${results}")"
#
#"%system.teamcity.build.checkoutDir%/tools/gh" pr comment "${pr_number}" --body "${body}"

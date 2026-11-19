#!/bin/bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Updates the body of the public "prioritized items" issue so that it lists every
# issue and pull request currently on the team working board. Draft items are
# excluded. A list item containing only an issue or pull request URL is rendered
# by GitHub with the item's title and state.
#
# Required environment:
#   ISSUE_NUMBER   - number of the issue in REPO whose body is updated
#   GH_TOKEN       - token used to read and edit the issue (issues: write)
#   PROJECT_TOKEN  - token with access to the working board
# Optional environment:
#   DRY_RUN        - "true" to log the new body without editing the issue

set -euo pipefail

PROJECT_NUMBER=196
REPO="hashicorp/terraform-provider-aws"
DRY_RUN="${DRY_RUN:-false}"

INTRO="Below is a list of items that the Terraform AWS Provider maintainers currently have prioritized. This list is regularly updated to provide insight as priorities shift and are updated."

# One URL per non-archived issue or pull request on the board. Drafts have no URL and are skipped.
board_urls () {
  GH_TOKEN="$PROJECT_TOKEN" gh api graphql --paginate -F project="$PROJECT_NUMBER" -f query='
    query($project: Int!, $endCursor: String) {
      organization(login: "hashicorp") {
        projectV2(number: $project) {
          items(first: 100, after: $endCursor) {
            nodes {
              isArchived
              content {
                ...on Issue { url }
                ...on PullRequest { url }
              }
            }
            pageInfo { endCursor hasNextPage }
          }
        }
      }
    }' --jq '.data.organization.projectV2.items.nodes[] | select(.isArchived | not) | .content.url // empty'
}

main () {
  local urls body current
  urls=$(board_urls)

  body="$INTRO"$'\n'
  if [[ -n "$urls" ]]; then
    body+=$'\n'$(sed 's/^/- /' <<<"$urls")
  fi

  # GitHub may store line endings as CRLF
  current=$(gh issue view "$ISSUE_NUMBER" --repo "$REPO" --json body --jq '.body' | tr -d '\r')

  if [[ "$current" == "$body" ]]; then
    echo "Issue #$ISSUE_NUMBER is already up to date ($(grep -c . <<<"$urls" || true) items)"
    return
  fi

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "[dry-run] Would update issue #$ISSUE_NUMBER with body:"
    echo "$body"
    return
  fi

  echo "Updating issue #$ISSUE_NUMBER"
  gh issue edit "$ISSUE_NUMBER" --repo "$REPO" --body "$body"
}

main

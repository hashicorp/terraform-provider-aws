#!/bin/bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Removes the prioritized label from every issue and pull request in a closed
# milestone, and from issues closed by those pull requests, then deletes them
# from the team working board.
#
# Required environment:
#   MILESTONE      - milestone number
#   GH_TOKEN       - token used to remove labels (GITHUB_TOKEN)
#   PROJECT_TOKEN  - token with access to the working board

set -euo pipefail

PROJECT_NUMBER=196
PROJECT_ID="PVT_kwDOAAuecM4AF-7h"

# Emits one JSON object per item: {type, url, prioritized, items}
item_fields='
  url
  labels(first: 100) { nodes { name } }
  projectItems(first: 20, includeArchived: true) { nodes { id project { id } } }
'

normalize='{
  url,
  prioritized: ([.labels.nodes[].name] | index("prioritized") != null),
  items: [.projectItems.nodes[] | select(.project.id == $project) | .id]
}'

milestone_issues () {
  GH_TOKEN="$PROJECT_TOKEN" gh api graphql --paginate -F milestone="$MILESTONE" -f query="
    query(\$milestone: Int!, \$endCursor: String) {
      repository(owner: \"hashicorp\", name: \"terraform-provider-aws\") {
        milestone(number: \$milestone) {
          issues(first: 50, after: \$endCursor) {
            nodes { $item_fields }
            pageInfo { endCursor hasNextPage }
          }
        }
      }
    }" --jq '.data.repository.milestone.issues.nodes[]' |
    jq --compact-output --arg project "$PROJECT_ID" "$normalize + {type: \"issue\"}"
}

milestone_pulls () {
  GH_TOKEN="$PROJECT_TOKEN" gh api graphql --paginate -F milestone="$MILESTONE" -f query="
    query(\$milestone: Int!, \$endCursor: String) {
      repository(owner: \"hashicorp\", name: \"terraform-provider-aws\") {
        milestone(number: \$milestone) {
          pullRequests(first: 25, after: \$endCursor) {
            nodes {
              $item_fields
              closingIssuesReferences(first: 25) { nodes { $item_fields } }
            }
            pageInfo { endCursor hasNextPage }
          }
        }
      }
    }" --jq '.data.repository.milestone.pullRequests.nodes[]' |
    jq --compact-output --arg project "$PROJECT_ID" "
      ($normalize + {type: \"pr\"}),
      (.closingIssuesReferences.nodes[] | $normalize + {type: \"issue\"})"
}

main () {
  local records
  records=$( { milestone_issues; milestone_pulls; } | jq --compact-output --slurp 'unique_by(.url) | .[]')

  if [[ -z "$records" ]]; then
    echo "No items found in milestone $MILESTONE"
    return
  fi

  # gh issue edit accepts multiple URLs at once, gh pr edit does not
  local issues
  issues=$(jq --raw-output --slurp '[.[] | select(.type == "issue" and .prioritized) | .url] | join(" ")' <<<"$records")
  if [[ -n "$issues" ]]; then
    # shellcheck disable=SC2086
    gh issue edit $issues --remove-label prioritized
  fi

  for pr in $(jq --raw-output 'select(.type == "pr" and .prioritized) | .url' <<<"$records"); do
    gh pr edit "$pr" --remove-label prioritized
  done

  for item in $(jq --raw-output '.items[]' <<<"$records"); do
    echo "Deleting $item"
    GH_TOKEN="$PROJECT_TOKEN" gh project item-delete "$PROJECT_NUMBER" --owner "hashicorp" --id "$item"
  done
}

main

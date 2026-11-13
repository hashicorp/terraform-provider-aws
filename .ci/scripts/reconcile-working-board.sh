#!/bin/bash
# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# Reconciles the team working board with the prioritized label
# Only items belonging to the repositories in REPOS are considered.
#
# Required environment:
#   GH_TOKEN  - token with access to the working board
# Optional environment:
#   DRY_RUN   - "true" to log actions without making changes

set -euo pipefail

PROJECT_NUMBER=196
PROJECT_ID="PVT_kwDOAAuecM4AF-7h"
VIEW_FIELD_ID="PVTSSF_lADOAAuecM4AF-7hzgMRB34"
WORKING_BOARD_VIEW_ID="8d366764"
REPOS=("hashicorp/terraform-provider-aws" "hashicorp/terraform-provider-awscc")
DRY_RUN="${DRY_RUN:-false}"

run () {
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

repos_json () {
  printf '%s\n' "${REPOS[@]}" | jq --raw-input --slurp --compact-output 'split("\n") | map(select(length > 0))'
}

# One JSON object per non-archived board item. Issues/pull requests are keyed as "owner/repo#number":
#   {id, type: "repo", key, view} for issues/pull requests in REPOS
#   {id, type: "draft", title, links: [keys]} for draft issues
#   {id, type: "other"} for anything else
board_items () {
  gh api graphql --paginate -f query='
    query($endCursor: String) {
      organization(login: "hashicorp") {
        projectV2(number: 196) {
          items(first: 100, after: $endCursor) {
            nodes {
              id
              isArchived
              fieldValueByName(name: "View") {
                ...on ProjectV2ItemFieldSingleSelectValue { name }
              }
              content {
                __typename
                ...on Issue { number repository { nameWithOwner } }
                ...on PullRequest { number repository { nameWithOwner } }
                ...on DraftIssue { title body }
              }
            }
            pageInfo { endCursor hasNextPage }
          }
        }
      }
    }' --jq '.data.organization.projectV2.items.nodes[] | select(.isArchived | not)' |
    jq --compact-output --argjson repos "$(repos_json)" '
      ("github\\.com/(" + ($repos | map(gsub("\\."; "\\.")) | join("|")) + ")/(?:issues|pull)/([0-9]+)") as $pattern
      | if (.content.__typename == "Issue" or .content.__typename == "PullRequest")
        and (.content.repository.nameWithOwner as $r | $repos | index($r)) then
        {id, type: "repo", key: "\(.content.repository.nameWithOwner)#\(.content.number)", view: .fieldValueByName.name}
      elif .content.__typename == "DraftIssue" then
        {
          id,
          type: "draft",
          title: .content.title,
          links: ([(.content.body // "") | scan($pattern) | "\(.[0])#\(.[1])"] | unique)
        }
      else
        {id, type: "other"}
      end'
}

# One "owner/repo#number" string per issue/pull request in REPOS with the prioritized label.
# Pass "open" to only include open items.
labeled_keys () {
  local query="label:prioritized"
  for repo in "${REPOS[@]}"; do
    query+=" repo:$repo"
  done
  if [[ "${1:-}" == "open" ]]; then
    query+=" is:open"
  fi
  gh api graphql --paginate -f searchQuery="$query" -f query='
    query($searchQuery: String!, $endCursor: String) {
      search(query: $searchQuery, type: ISSUE, first: 100, after: $endCursor) {
        nodes {
          ...on Issue { number repository { nameWithOwner } }
          ...on PullRequest { number repository { nameWithOwner } }
        }
        pageInfo { endCursor hasNextPage }
      }
    }' --jq '.data.search.nodes[] | "\(.repository.nameWithOwner)#\(.number)"'
}

main () {
  local items labeled open_labeled
  items=$(board_items | jq --compact-output --slurp '.')
  labeled=$(labeled_keys | jq --raw-input --compact-output --slurp 'split("\n") | map(select(length > 0)) | unique')
  open_labeled=$(labeled_keys open | jq --raw-input --compact-output --slurp 'split("\n") | map(select(length > 0)) | unique')

  echo "Repos: ${REPOS[*]}"
  echo "Board items: $(jq 'length' <<<"$items"), labeled items: $(jq 'length' <<<"$labeled")"

  # Delete issues/pull requests on the board without the prioritized label
  for item in $(jq --raw-output --argjson labeled "$labeled" '
      .[] | select(.type == "repo" and (.key as $k | $labeled | index($k) | not))
      | "\(.id),\(.key)"' <<<"$items"); do
    echo "Deleting unlabeled ${item#*,} (${item%%,*})"
    run gh project item-delete "$PROJECT_NUMBER" --owner hashicorp --id "${item%%,*}"
  done

  # Delete drafts with unlabeled links
  for id in $(jq --raw-output --argjson labeled "$labeled" '
      .[] | select(.type == "draft" and (.links | length > 0)
        and all(.links[]; . as $k | $labeled | index($k) | not))
      | .id' <<<"$items"); do
    echo "Deleting draft $id ($(jq --raw-output --arg id "$id" '.[] | select(.id == $id) | .title' <<<"$items"))"
    run gh project item-delete "$PROJECT_NUMBER" --owner hashicorp --id "$id"
  done

  # Adds any issues marked as prioritized to the board if they don't already exist there.
  for key in $(jq --raw-output --argjson labeled "$open_labeled" '
      ([.[] | select(.type == "repo") | .key] + [.[] | select(.type == "draft") | .links[]]) as $present
      | $labeled[] | select(. as $k | $present | index($k) | not)' <<<"$items"); do
    # GitHub redirects /issues/N to /pull/N for pull requests
    url="https://github.com/${key%#*}/issues/${key##*#}"
    echo "Adding labeled $key"
    if [[ "$DRY_RUN" == "true" ]]; then
      echo "[dry-run] gh project item-add $PROJECT_NUMBER --owner hashicorp --url $url"
      continue
    fi
    id=$(gh project item-add "$PROJECT_NUMBER" --owner hashicorp --url "$url" --format json --jq '.id')
    gh project item-edit \
      --id "$id" \
      --project-id "$PROJECT_ID" \
      --field-id "$VIEW_FIELD_ID" \
      --single-select-option-id "$WORKING_BOARD_VIEW_ID"
  done

  # Add to correct view if they are labeled without a view assigned.
  for item in $(jq --raw-output --argjson labeled "$labeled" '
      .[] | select(.type == "repo" and .view != "working-board" and (.key as $k | $labeled | index($k)))
      | "\(.id),\(.key)"' <<<"$items"); do
    echo "Setting View to working-board on ${item#*,} (${item%%,*})"
    run gh project item-edit \
      --id "${item%%,*}" \
      --project-id "$PROJECT_ID" \
      --field-id "$VIEW_FIELD_ID" \
      --single-select-option-id "$WORKING_BOARD_VIEW_ID"
  done
}

main

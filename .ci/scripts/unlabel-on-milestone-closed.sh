#!/bin/bash

main () {
  ISSUES=$(gh api graphql --paginate -F milestone="$MILESTONE" -f query='
    query($milestone: Int!, $endCursor: String) {
      organization(login: "hashicorp") {
        repository(name: "terraform-provider-aws") {
          milestone(number: $milestone) {
            issues(first: 10, after: $endCursor, labels: ["prioritized"]) {
              edges {
                node {
                  url
                }
              }
              pageInfo {
                endCursor
                hasNextPage
              }
            }
          }
        }
      }
    }' --jq '.data.organization.repository.milestone.issues.edges[].node' | jq --raw-output --slurp '[.[].url] | join(" ")')

  PULLS=$(gh api graphql --paginate -F milestone="$MILESTONE" -f query='
    query($milestone: Int!, $endCursor: String) {
      organization(login: "hashicorp") {
        repository(name: "terraform-provider-aws") {
          milestone(number: $milestone) {
            pullRequests(first: 10, after: $endCursor, labels: ["prioritized"]) {
              edges {
                node {
                  url
                }
              }
              pageInfo {
                endCursor
                hasNextPage
              }
            }
          }
        }
      }
    }' --jq '.data.organization.repository.milestone.pullRequests.edges[].node' | jq --raw-output --slurp '.[] | .url')

  # gh issues allows passing multiple URLs at once
  gh issue edit $ISSUES --remove-label prioritized

  # gh pr does not allow passing multiple URLs at once
  for item in $PULLS; do
    gh pr edit "$item" --remove-label prioritized
  done
}

main

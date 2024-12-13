#!/bin/bash

PROJECT_NUMBER=196
PROJECT_ID="PVT_kwDOAAuecM4AF-7h"

main () {
  ISSUES=$(gh api graphql --paginate -F milestone="$MILESTONE" -f query='
    query($milestone: Int!, $endCursor: String) {
      organization(login: "hashicorp") {
        repository(name: "terraform-provider-aws") {
          milestone(number: $milestone) {
            issues(first: 10, after: $endCursor) {
              edges {
                node {
                  id
                  projectItems(includeArchived: false, first: 10) {
                    nodes {
                      id
                      project {
                        id
                      }
                    }
                  }
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
    }' --jq '.data.organization.repository.milestone.issues.edges[].node.projectItems.nodes[]' | jq --slurp '. | tojson')

  PULLS=$(gh api graphql --paginate -F milestone="$MILESTONE" -f query='
    query($milestone: Int!, $endCursor: String) {
      organization(login: "hashicorp") {
        repository(name: "terraform-provider-aws") {
          milestone(number: $milestone) {
            pullRequests(first: 10, after: $endCursor) {
              edges {
                node {
                  id
                  projectItems(includeArchived: false, first: 10) {
                    nodes {
                      id
                      project {
                        id
                      }
                    }
                  }
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
    }' --jq '.data.organization.repository.milestone.pullRequests.edges[].node.projectItems.nodes[]' | jq --slurp '. | tojson')

  PROJECT_ITEMS=$(jq \
    --null-input \
    --arg project "$PROJECT_ID" \
    --argjson issues "$ISSUES" \
    --argjson pulls "$PULLS" \
    '$issues, $pulls | fromjson | .[] | select(.project.id == $project).id')

  for item in $PROJECT_ITEMS; do
    echo "Archiving $item"
    gh project item-archive "$PROJECT_NUMBER" --owner "hashicorp" --id "$item"
  done
}

main

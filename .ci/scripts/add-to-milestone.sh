#!/bin/bash

MILESTONE_NAME=$(gh api graphql -F current="$MILESTONE" -f query='
query($current: String!) {
  repository(name: "terraform-provider-aws", owner: "hashicorp") {
    milestones(query: $current, first: 1) {
      nodes {
        title
      }
    }
  }
}' --jq '.data.repository.milestones.nodes[].title')

# Get the issues related to this pull request and
DATA=$(gh api graphql --paginate -F number="$PR_NUMBER" -f query='
query($number: Int!, $endCursor: String) {
  repository(name: "terraform-provider-aws", owner: "hashicorp") {
    pullRequest(number: $number) {
      id
      url
      closingIssuesReferences(first: 1, after: $endCursor) {
        pageInfo {
          endCursor
          hasNextPage
        }
        edges {
          node {
            url
            state
            timelineItems(itemTypes: CLOSED_EVENT, last: 3) {
              nodes {
                ... on ClosedEvent {
                  closer {
                    ... on PullRequest {
                      id
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}' | jq --slurp '. | tojson')

# Due to --slurp, we'll have a list of multiple objects starting at .data
# but we only need the PR ID once. Needed for comparing against the closer
# on the related issues.
PULL_ID=$(jq \
  --null-input \
  --argjson data "$DATA" \
  '$data | fromjson | .[0].data.repository.pullRequest.id')

# Similar comment to above, but we need the URL for the gh call.
PULL_URL=$(jq \
  --null-input \
  --raw-output \
  --argjson data "$DATA" \
  '$data | fromjson | .[0].data.repository.pullRequest.url')

# Get the URLs for all issues closed by this pull request. Needed for the gh call.
ISSUES_URLS=($(jq \
  --null-input \
  --raw-output \
  --argjson data "$DATA" \
  --argjson pull "$PULL_ID" \
  '$data
  | fromjson
  | .[].data.repository.pullRequest.closingIssuesReferences.edges[]
  | select(.node.state == "CLOSED")
  | select(.node.timelineItems.nodes[].closer.id == $pull).node.url'))


# Add pull request to milestone
gh pr edit "$PULL_URL" --milestone "$MILESTONE_NAME"

# Add issues to milestone
gh issue edit "${ISSUES_URLS[@]}" --milestone "$MILESTONE_NAME"

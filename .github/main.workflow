workflow "Pull Request Updates" {
    on       = "pull_request"
    resolves = "pr-label"
}

action "pr-filter-sync" {
    uses = "actions/bin/filter@master"
    args = "action 'opened|synchronize'"
}

action "pr-label" {
    uses    = "actions/labeler@v1.0.0"
    needs   = "pr-filter-sync"

    secrets = [
        "GITHUB_TOKEN"
    ]

    env     = {
        LABEL_SPEC_FILE = ".github/PULL_REQUEST_LABELS.yml"
    }
}

workflow "Issue triage" {
  on = "issues"
  resolves = "Apply Triage Label"
}

action "Apply Triage Label" {
  uses = "actions/github@v1.0.0"
  args = "label needs-triage --action=opened" # Only when issues are opened!
  secrets = ["GITHUB_TOKEN"]
}

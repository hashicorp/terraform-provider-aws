variable "workflow_labels" {
  default = {
    "provider"                       = "623ce4",
    "needs-triage"                   = "ca2171",
    "enhancement"                    = "962fab",
    "new-resource"                   = "962fab",
    "new-data-source"                = "962fab",
    "bug"                            = "f04e54",
    "crash"                          = "f04e54",
    "breaking-change"                = "f04e54",
    "regression"                     = "f04e54",
    "waiting-response"               = "ef4349",
    "tests"                          = "00bc7f",
    "terraform-0.11"                 = "00bc7f",
    "terraform-0.12"                 = "00bc7f",
    "terraform-0.13"                 = "00bc7f",
    "terraform-0.14"                 = "00bc7f",
    "terraform-0.15"                 = "00bc7f",
    "alpha-testing"                  = "00bc7f",
    "documentation"                  = "b0288e",
    "technical-debt"                 = "b0288e",
    "proposal"                       = "b0288e",
    "thinking"                       = "b0288e",
    "question"                       = "b0288e",
    "linter"                         = "b0288e",
    "size/XS"                        = "14c6cb",
    "size/S"                         = "12b4ba",
    "size/M"                         = "11a2a7",
    "size/L"                         = "0f9095",
    "size/XL"                        = "0d7e82",
    "size/XXL"                       = "0b6c6f",
    "upstream-terraform"             = "1563ff",
    "upstream"                       = "1563ff",
    "go"                             = "1563ff",
    "dependencies"                   = "1563ff",
    "good first issue"               = "00acff",
    "help wanted"                    = "00acff",
    "examples"                       = "00acff",
    "hacktoberfest"                  = "828a90",
    "stale"                          = "828a90",
    "hashibot/ignore"                = "828a90",
    "new"                            = "828a90",
    "windows"                        = "828a90",
    "reinvent"                       = "828a90",
    "github_actions"                 = "828a90",
    "terraform-plugin-sdk-migration" = "828a90",
    "terraform-plugin-sdk-v1"        = "828a90",
  }
  description = "Name-color mapping of workflow issues"
  type        = map(string)
}

resource "github_issue_label" "workflow" {
  for_each = var.workflow_labels

  repository = "terraform-provider-aws"
  name       = each.key
  color      = each.value
}

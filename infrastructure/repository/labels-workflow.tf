variable "workflow_labels" {
  default = {
    "provider" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to the provider itself, rather than any interaction with AWS.",
    },
    "timeouts" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to timeout increases."
    },
    "eventual-consistency" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to eventual consistency issues."
    },
    "tags" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to resource tagging."
    },
    "authentication" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to authentication; to the provider itself of otherwise."
    },
    "fips" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to the Federal Information Processing Standard (FIPS)."
    },
    "needs-triage" = {
      color       = "dc477d", # color:consul
      description = "Waiting for first response or review from a maintainer."
    },
    "enhancement" = {
      color       = "844fba", # color:terraform (main)
      description = "Requests to existing resources that expand the functionality or scope."
    },
    "new-resource" = {
      color       = "8040c9", # color:terraform (link on white)
      description = "Introduces a new resource"
    },
    "new-data-source" = {
      color       = "ac72f0", # color:terraform (link on black)
      description = "Introduces a new data source."
    },
    "bug" = {
      color       = "ec585d", # color:boundary
      description = "Addresses a defect in current functionality."
    },
    "crash" = {
      color       = "ec585d", # color:boundary
      description = "Results from or addresses a Terraform crash or kernel panic."
    },
    "breaking-change" = {
      color       = "ec585d", # color:boundary
      description = "Introduces a breaking change in current functionality; usually deferred to the next major release."
    },
    "regression" = {
      color       = "ec585d", # color:boundary
      description = "Pertains to a degraded workflow resulting from an upstream patch or internal enhancement."
    },
    "waiting-response" = {
      color       = "d3353f", # color:darker boundary
      description = "Maintainers are waiting on response from community or contributor."
    },
    "tests" = {
      color       = "60dea9", # color:nomad
      description = "PRs: expanded test coverage. Issues: expanded coverage, enhancements to test infrastructure."
    },
    "prerelease-tf-testing" = {
      color       = "60dea9", # color:nomad
      description = "Pertains to testing Terraform releases prior to release."
    },
    "technical-debt" = {
      color       = "d1ebff", # color:terraform accent
      description = "Addresses areas of the codebase that need refactoring or redesign."
    },
    "proposal" = {
      color       = "d1ebff", # color:terraform accent
      description = "Proposes new design or functionality."
    },
    "documentation" = {
      color       = "f4ecff", # color:terraform secondary
      description = "Introduces or discusses updates to documentation."
    },
    "question" = {
      color       = "f4ecff", # color:terraform secondary
      description = "A question about existing functionality; most questions are re-routed to discuss.hashicorp.com."
    },
    "linter" = {
      color       = "f4ecff", # color:terraform secondary
      description = "Pertains to changes to or issues with the various linters."
    },
    "sweeper" = {
      color       = "f4ecff", # color:terraform secondary
      description = "Pertains to changes to or issues with the sweeper."
    },
    "size/XS" = {
      color       = "62d4dc", # color:lightest-darkest waypoint gradient
      description = "Managed by automation to categorize the size of a PR."
    },
    "size/S" = {
      color       = "4ec3ce", # color:lightest-darkest waypoint gradient
      description = "Managed by automation to categorize the size of a PR."
    },
    "size/M" = {
      color       = "3bb3c0", # color:lightest-darkest waypoint gradient
      description = "Managed by automation to categorize the size of a PR."
    },
    "size/L" = {
      color       = "27a2b2", # color:lightest-darkest waypoint gradient
      description = "Managed by automation to categorize the size of a PR."
    },
    "size/XL" = {
      color       = "1492a4", # color:lightest-darkest waypoint gradient
      description = "Managed by automation to categorize the size of a PR."
    },
    "size/XXL" = {
      color       = "008196", # color:lightest-darkest waypoint gradient
      description = "Managed by automation to categorize the size of a PR."
    },
    "upstream-terraform" = {
      color       = "1c7ada", # color:vagrant
      description = "Addresses functionality related to the Terraform core binary."
    },
    "upstream" = {
      color       = "1c7ada", # color:vagrant
      description = "Addresses functionality related to the cloud provider."
    },
    "dependencies" = {
      color       = "1c7ada", # color:vagrant
      description = "Used to indicate dependency changes."
    },
    "good first issue" = {
      color       = "63d0ff", # color:packer
      description = "Call to action for new contributors looking for a place to start. Smaller or straightforward issues."
    },
    "examples" = {
      color       = "63d0ff", # color:packer
      description = "Introduces or discusses updates to examples."
    },
    "stale" = {
      color       = "828a90", # color:stale grey
      description = "Old or inactive issues managed by automation, if no further action taken these will get closed."
    },
    "windows" = {
      color       = "828a90", # color:stale grey
      description = "Issues and PRs that relate to using the provider on the Windows operating system."
    },
    "repository" = {
      color       = "828a90", # color:stale grey
      description = "Repository modifications; GitHub Actions, developer docs, issue templates, codeowners, changelog."
    },
  }
  description = "Name-color-description mapping of workflow issues."
  type        = map(any)
}

resource "github_issue_label" "workflow" {
  for_each = var.workflow_labels

  repository  = "terraform-provider-aws"
  name        = each.key
  color       = each.value.color
  description = each.value.description
}

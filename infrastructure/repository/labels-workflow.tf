variable "workflow_labels" {
  default = {
    # Stop the search. Keep these alphabetibelized.

    "authentication" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to authentication; to the provider itself of otherwise."
    },
    "breaking-change" = {
      color       = "ec585d", # color:boundary
      description = "Introduces a breaking change in current functionality; usually deferred to the next major release."
    },
    "bug" = {
      color       = "ec585d", # color:boundary
      description = "Addresses a defect in current functionality."
    },
    "client-connections" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to the AWS Client and service connections."
    },
    "crash" = {
      color       = "ec585d", # color:boundary
      description = "Results from or addresses a Terraform crash or kernel panic."
    },
    "create" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to generating names, hashcodes, etc."
    },
    "dependencies" = {
      color       = "1c7ada", # color:vagrant
      description = "Used to indicate dependency changes."
    },
    "documentation" = {
      color       = "f4ecff", # color:terraform secondary
      description = "Introduces or discusses updates to documentation."
    },
    "enhancement" = {
      color       = "844fba", # color:terraform (main)
      description = "Requests to existing resources that expand the functionality or scope."
    },
    "eventual-consistency" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to eventual consistency issues."
    },
    "examples" = {
      color       = "63d0ff", # color:packer
      description = "Introduces or discusses updates to examples."
    },
    "flex" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to FLatteners and EXpanders."
    },
    "fips" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to the Federal Information Processing Standard (FIPS)."
    },
    "generators" = {
      color       = "60dea9", # color:nomad
      description = "Relates to code generators."
    },
    "good first issue" = {
      color       = "63d0ff", # color:packer
      description = "Call to action for new contributors looking for a place to start. Smaller or straightforward issues."
    },
    "linter" = {
      color       = "f4ecff", # color:terraform secondary
      description = "Pertains to changes to or issues with the various linters."
    },
    "needs-triage" = {
      color       = "dc477d", # color:consul
      description = "Waiting for first response or review from a maintainer."
    },
    "new-data-source" = {
      color       = "ac72f0", # color:terraform (link on black)
      description = "Introduces a new data source."
    },
    "new-resource" = {
      color       = "8040c9", # color:terraform (link on white)
      description = "Introduces a new resource."
    },
    "new-service" = {
      color       = "ac72f0", # color:terraform (link on black)
      description = "Introduces a new service."
    },
    "pre-service-packages" = {
      color       = "ffec6e", # color:vault
      description = "Includes pre-Service Packages aspects."
    },
    "prerelease-tf-testing" = {
      color       = "60dea9", # color:nomad
      description = "Pertains to testing Terraform releases prior to release."
    },
    "proposal" = {
      color       = "d1ebff", # color:terraform accent
      description = "Proposes new design or functionality."
    },
    "provider" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to the provider itself, rather than any interaction with AWS.",
    },
    "question" = {
      color       = "f4ecff", # color:terraform secondary
      description = "A question about existing functionality; most questions are re-routed to discuss.hashicorp.com."
    },
    "regression" = {
      color       = "ec585d", # color:boundary
      description = "Pertains to a degraded workflow resulting from an upstream patch or internal enhancement."
    },
    "repository" = {
      color       = "828a90", # color:stale grey
      description = "Repository modifications; GitHub Actions, developer docs, issue templates, codeowners, changelog."
    },
    "service/meta" = {
      color       = "7b42bc", # color:terraform (logomark)
      description = "Issues and PRs that correspond to meta data sources."
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
    "skaff" = {
      color       = "63d0ff", # color:packer
      description = "Issues and pull requested related to the skaff tool"
    }
    "stale" = {
      color       = "828a90", # color:stale grey
      description = "Old or inactive issues managed by automation, if no further action taken these will get closed."
    },
    "sweeper" = {
      color       = "f4ecff", # color:terraform secondary
      description = "Pertains to changes to or issues with the sweeper."
    },
    "tags" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to resource tagging."
    },
    "technical-debt" = {
      color       = "d1ebff", # color:terraform accent
      description = "Addresses areas of the codebase that need refactoring or redesign."
    },
    "tests" = {
      color       = "60dea9", # color:nomad
      description = "PRs: expanded test coverage. Issues: expanded coverage, enhancements to test infrastructure."
    },
    "timeouts" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to timeout increases."
    },
    "upstream" = {
      color       = "1c7ada", # color:vagrant
      description = "Addresses functionality related to the cloud provider."
    },
    "upstream-terraform" = {
      color       = "1c7ada", # color:vagrant
      description = "Addresses functionality related to the Terraform core binary."
    },
    "verify" = {
      color       = "844fba", # color:terraform (main)
      description = "Pertains to the verify package (i.e., provider-level validating, diff suppression, etc.)"
    },
    "waiting-response" = {
      color       = "d3353f", # color:darker boundary
      description = "Maintainers are waiting on response from community or contributor."
    },
    "windows" = {
      color       = "828a90", # color:stale grey
      description = "Issues and PRs that relate to using the provider on the Windows operating system."
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

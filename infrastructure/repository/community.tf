// A list of maintainers to be used as an "allow list" for various GitHub Actions.
// This allows us to make various "exceptions" for maintainers, such as automatically
// removing the `needs-triage` label from new Issues and Pull Requests
//
resource "github_actions_secret" "maintainer_list" {
  repository      = "terraform-provider-aws"
  secret_name     = "MAINTAINER_LIST"
  plaintext_value = "['breathingdust', 'dependabot[bot]', 'ewbankkit', 'gdavison', 'jar-b', 'johnsonaj', 'justinretzolk', 'marcosentino', 'nam054', 'YakDriver']"
}

variable "community_list_repo" {
  type        = string
  description = "The name of the repository containing the lists of users."
}

// Core Contributors
data "github_repository_file" "core_contributors" {
  repository = var.community_list_repo
  file       = "core-contributors.json"
}

resource "github_actions_secret" "core_contributors" {
  repository      = "terraform-provider-aws"
  secret_name     = "CORE_CONTRIBUTORS"
  plaintext_value = base64encode(data.github_repository_file.core_contributors.content)
}

// Maintainers
data "github_team" "aws" {
  slug = "terraform-aws"
}

resource "github_actions_secret" "maintainers" {
  repository      = "terraform-provider-aws"
  secret_name     = "MAINTAINERS"
  plaintext_value = base64encode(jsonencode(data.github_team.aws.members))
}

// Partners
data "github_repository_file" "partners" {
  repository = var.community_list_repo
  file       = "partners.json"
}

resource "github_actions_secret" "partners" {
  repository      = "terraform-provider-aws"
  secret_name     = "PARTNERS"
  plaintext_value = base64encode(data.github_repository_file.partners.content)
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "community_list_repo" {
  type        = string
  description = "The name of the repository containing the lists of users. Value set in TFC."
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
  plaintext_value = base64encode(jsonencode(concat(data.github_team.aws.members, ["dependabot[bot]"])))
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

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
  ignore_tags {
    keys = var.ignore_tag_keys
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_mpa_approval_team" "test" {
  name        = var.rName
  description = "Test approval team"

  approval_strategy {
    m_of_n {
      min_approvals_required = 1
    }
  }

  approver {
    primary_identity_id         = data.aws_caller_identity.current.user_id
    primary_identity_source_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
  }

  policy {
    policy_arn = "arn:aws:mpa:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:policy/example"
  }

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = true
  default  = null
}

variable "ignore_tag_keys" {
  type     = set(string)
  nullable = false
}

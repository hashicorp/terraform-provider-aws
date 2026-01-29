# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

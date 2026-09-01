# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_arcregionswitch_plan" "test" {
  name              = var.rName
  execution_role    = aws_iam_role.test.arn
  recovery_approach = "activePassive"
  regions           = [local.primary_region, local.secondary_region]
  primary_region    = local.primary_region

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = local.secondary_region
    step {
      name                 = "basic-step"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  workflow {
    workflow_target_action = "activate"
    workflow_target_region = local.primary_region

    step {
      name                 = "basic-step-primary"
      execution_block_type = "ManualApproval"

      execution_approval_config {
        approval_role   = aws_iam_role.test.arn
        timeout_minutes = 60
      }
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "arc-region-switch.amazonaws.com"
        }
      },
    ]
  })
}

locals {
  primary_region   = data.aws_region.current.name
  secondary_region = data.aws_region.secondary.name
}

data "aws_region" "current" {}

data "aws_region" "secondary" {
  region = var.secondary_region
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

variable "secondary_region" {
  description = "Secondary region"
  type        = string
  nullable    = false
}

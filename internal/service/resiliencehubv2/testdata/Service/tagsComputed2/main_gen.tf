# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

data "aws_region" "current" {
}

resource "aws_resiliencehubv2_policy" "test" {
  name = "${var.rName}-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  name    = var.rName
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
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

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}

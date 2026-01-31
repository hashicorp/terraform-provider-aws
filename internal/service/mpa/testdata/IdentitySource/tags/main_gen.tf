# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {}
data "aws_region" "current" {}

resource "aws_mpa_identity_source" "test" {
  name = var.rName

  identity_source_parameters {
    iam_identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      region       = data.aws_region.current.name
    }
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

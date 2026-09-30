# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_accountaccess_application" "test" {
  identity_source {
    identity_center {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    }
  }

  tags = var.resource_tags
}

data "aws_ssoadmin_instances" "test" {}

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}

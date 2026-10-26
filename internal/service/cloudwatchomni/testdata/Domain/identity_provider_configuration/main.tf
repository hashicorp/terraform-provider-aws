# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_domain" "test" {
  name               = var.rName
  identity_providers = ["IAM", "IDC"]

  identity_provider_configuration {
    identity_center_configuration {
      identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    }
  }
}

data "aws_ssoadmin_instances" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

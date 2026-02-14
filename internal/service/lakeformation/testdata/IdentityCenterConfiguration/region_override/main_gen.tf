# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lakeformation_identity_center_configuration" "test" {
  region = var.region

  instance_arn = local.identity_center_instance_arn
}

locals {
  identity_center_instance_arn = data.aws_ssoadmin_instances.test.arns[0]
}

data "aws_ssoadmin_instances" "test" {}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

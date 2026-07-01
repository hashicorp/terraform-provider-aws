# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

resource "aws_ssoadmin_region" "test" {
  region = var.region

  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  region_name  = "us-west-2"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

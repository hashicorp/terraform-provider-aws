# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_ssoadmin_instances" "test" {
  region = var.region
}

resource "aws_ssoadmin_region" "test" {
  count  = var.resource_count
  region = var.region

  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  region_name  = element(var.region_names, count.index)
}

variable "region_names" {
  description = "Region names to enable on the instance"
  type        = list(string)
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambda_layer_version" "test" {
  count  = var.resource_count
  region = var.region

  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "${var.rName}-${count.index}"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to create resources in"
  type        = string
  nullable    = false
}

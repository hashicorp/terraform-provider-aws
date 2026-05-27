# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_cloudwatch_log_group" "test" {
  count = 3

  region = var.region

  name = "${var.rName}-${count.index}"

  retention_in_days = 1
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

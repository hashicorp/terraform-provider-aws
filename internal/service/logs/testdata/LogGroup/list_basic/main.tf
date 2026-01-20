# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_cloudwatch_log_group" "test" {
  count = 3

  name = "${var.rName}-${count.index}"

  retention_in_days = 1
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

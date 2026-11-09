# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kinesis_stream" "test" {
  name        = var.rName
  shard_count = 2
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

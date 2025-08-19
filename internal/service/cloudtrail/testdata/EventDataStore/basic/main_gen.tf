# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudtrail_event_data_store" "test" {
  name = var.rName

  termination_protection_enabled = false # For ease of deletion.
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

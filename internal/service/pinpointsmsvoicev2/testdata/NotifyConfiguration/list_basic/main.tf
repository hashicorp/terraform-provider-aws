# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  count = var.resource_count

  display_name     = "${var.rName}-${count.index}"
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS"]
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

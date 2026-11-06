# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  display_name     = var.rName
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

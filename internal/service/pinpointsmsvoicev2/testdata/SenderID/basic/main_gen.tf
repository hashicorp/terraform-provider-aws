# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = var.rName
  iso_country_code = "GB"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

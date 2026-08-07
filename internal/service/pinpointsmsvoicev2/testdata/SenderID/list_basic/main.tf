# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  count = var.resource_count

  sender_id        = "${var.rName}${count.index}"
  iso_country_code = "GB"
}

variable "rName" {
  description = "Base name for the sender IDs, suffixed with the resource index"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  count = var.resource_count

  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = upper("${var.rName}-${count.index}")
  keyword_message      = "test keyword message"
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

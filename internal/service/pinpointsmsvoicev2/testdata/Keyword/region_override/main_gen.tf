# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  region = var.region

  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = var.rName
  keyword_message      = "test keyword message"
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  region = var.region

  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
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

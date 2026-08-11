# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_pool" "test" {
  region = var.region

  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  region = var.region

  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

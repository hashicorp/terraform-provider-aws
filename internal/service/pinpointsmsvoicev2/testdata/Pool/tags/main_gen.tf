# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]

  tags = var.resource_tags
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

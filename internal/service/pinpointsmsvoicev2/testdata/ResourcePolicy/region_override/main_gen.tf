# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_resource_policy" "test" {
  region = var.region

  resource_arn = aws_pinpointsmsvoicev2_phone_number.test.arn
  policy       = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["sms-voice:SendTextMessage"]
    resources = [aws_pinpointsmsvoicev2_phone_number.test.arn]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  region = var.region

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

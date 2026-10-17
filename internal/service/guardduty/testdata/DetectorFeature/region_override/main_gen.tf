# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_guardduty_detector" "test" {
  region = var.region

  enable = true
}

resource "aws_guardduty_detector_feature" "test" {
  region = var.region

  detector_id = aws_guardduty_detector.test.id
  name        = "RDS_LOGIN_EVENTS"
  status      = "ENABLED"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

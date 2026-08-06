# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_guardduty_detector" "test" {
  enable = true
}

resource "aws_guardduty_detector_feature" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "RDS_LOGIN_EVENTS"
  status      = "ENABLED"
}


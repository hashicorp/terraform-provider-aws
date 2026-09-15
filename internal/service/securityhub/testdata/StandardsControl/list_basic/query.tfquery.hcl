# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_securityhub_standards_control" "test" {
  provider = aws

  config {
    standards_subscription_arn = aws_securityhub_standards_subscription.test.arn
  }
}

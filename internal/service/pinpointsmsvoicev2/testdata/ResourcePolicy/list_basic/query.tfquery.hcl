# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_pinpointsmsvoicev2_resource_policy" "test" {
  provider = aws

  config {
    resource_arn = aws_pinpointsmsvoicev2_phone_number.test.arn
  }
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_pinpointsmsvoicev2_keyword" "test" {
  provider = aws

  include_resource = true

  config {
    origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  }
}

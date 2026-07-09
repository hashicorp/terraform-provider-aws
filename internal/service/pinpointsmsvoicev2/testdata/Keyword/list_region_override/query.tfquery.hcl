# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_pinpointsmsvoicev2_keyword" "test" {
  provider = aws

  config {
    region = var.region
  }
}

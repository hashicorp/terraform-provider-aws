# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_wafv2_ip_set" "test" {
  provider = aws

  include_resource = true

  config {
    scope = "REGIONAL"
  }
}

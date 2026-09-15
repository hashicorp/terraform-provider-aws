# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_wafv2_web_acl_rule" "test" {
  provider = aws

  config {
    region      = var.region
    web_acl_arn = aws_wafv2_web_acl.test.arn
  }
}

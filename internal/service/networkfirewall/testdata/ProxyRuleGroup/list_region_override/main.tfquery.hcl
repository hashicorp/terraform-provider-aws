# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_networkfirewall_proxy_rule_group" "test" {
  provider = aws

  config {
    region = var.region
  }
}

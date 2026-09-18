# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ec2_network_insights_access_scope" "test" {
  provider = aws

  config {
    region = var.region
  }
}

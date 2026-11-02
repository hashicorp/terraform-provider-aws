# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_mailmanager_ingress_point" "test" {
  provider = aws

  config {
    region = var.region
  }
}

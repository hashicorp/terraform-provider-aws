# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_workmail_group" "test" {
  provider = aws

  config {
    organization_id = aws_workmail_organization.test.organization_id
    region          = var.region
  }
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_workmail_domain" "test" {
  provider = aws

  include_resource = true

  config {
    organization_id = aws_workmail_organization.test.organization_id
  }
}

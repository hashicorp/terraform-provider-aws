# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ssoadmin_application_grant" "test" {
  provider = aws

  include_resource = true

  config {
    application_arn = aws_ssoadmin_application.test.application_arn
  }
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ssoadmin_application_grant" "test" {
  provider = aws

  config {
    region          = var.region
    application_arn = aws_ssoadmin_application.test.application_arn
  }
}

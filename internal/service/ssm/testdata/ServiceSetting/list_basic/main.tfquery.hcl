# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ssm_service_setting" "test" {
  provider = aws

  config {
    arn = aws_ssm_service_setting.test.arn
  }
}

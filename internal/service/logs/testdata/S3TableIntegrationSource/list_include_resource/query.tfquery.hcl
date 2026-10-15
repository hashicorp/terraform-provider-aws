# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_log_s3_table_integration_source" "test" {
  provider = aws

  include_resource = true

  config {
    integration_arn = aws_observabilityadmin_s3_table_integration.test.arn
  }
}

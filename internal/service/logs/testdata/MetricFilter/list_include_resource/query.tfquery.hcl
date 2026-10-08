# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_cloudwatch_log_metric_filter" "test" {
  provider = aws

  include_resource = true

  config {
    log_group_name = aws_cloudwatch_log_group.test.name
  }
}

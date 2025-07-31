# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cognito_log_delivery_configuration" "test" {
  user_pool_id = aws_cognito_user_pool.test.id

  log_configurations {
    event_source = "userNotification"
    log_level    = "ERROR"

    cloud_watch_logs_configuration {
      log_group_arn = aws_cloudwatch_log_group.test.arn
    }
  }
}

resource "aws_cognito_user_pool" "test" {
  name = var.rName
}

resource "aws_cloudwatch_log_group" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cognito_log_delivery_configuration" "test" {
  region = var.region

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
  region = var.region

  name = var.rName
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

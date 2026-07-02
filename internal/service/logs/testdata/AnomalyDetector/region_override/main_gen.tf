# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_anomaly_detector" "test" {
  region = var.region

  detector_name           = var.rName
  log_group_arn_list      = [aws_cloudwatch_log_group.test.arn]
  anomaly_visibility_time = 7
  evaluation_frequency    = "TEN_MIN"
  enabled                 = "false"
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

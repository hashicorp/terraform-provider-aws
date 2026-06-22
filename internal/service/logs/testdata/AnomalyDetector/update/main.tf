# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_anomaly_detector" "test" {
  detector_name           = var.rName
  log_group_arn_list      = [aws_cloudwatch_log_group.test.arn]
  anomaly_visibility_time = var.anomalyVisibilityTime
  evaluation_frequency    = var.evaluationFrequency
  enabled                 = var.enabled
}

resource "aws_cloudwatch_log_group" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "evaluationFrequency" {
  type     = string
  nullable = false
}

variable "enabled" {
  type     = bool
  nullable = false
}

variable "anomalyVisibilityTime" {
  type     = number
  nullable = false
}

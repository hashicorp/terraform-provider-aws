# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_metric_filter" "test" {
  count  = var.resource_count
  region = var.region

  name           = "${var.rName}-${count.index}"
  pattern        = ""
  log_group_name = aws_cloudwatch_log_group.test.name

  metric_transformation {
    name      = "metric1"
    namespace = "ns1"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name              = "${var.rName}-group"
  retention_in_days = 1
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

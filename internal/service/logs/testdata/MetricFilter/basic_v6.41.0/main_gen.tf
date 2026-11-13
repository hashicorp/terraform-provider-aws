# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_metric_filter" "test" {
  name           = "${var.rName}-filter"
  pattern        = ""
  log_group_name = aws_cloudwatch_log_group.test.name

  metric_transformation {
    name      = "metric1"
    namespace = "ns1"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name              = "${var.rName}-group"
  retention_in_days = 1
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.41.0"
    }
  }
}

provider "aws" {}

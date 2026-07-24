# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_wafv2_web_acl" "test" {
  region = var.region
  name   = var.rName
  scope  = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = var.rName
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test0" {
  region      = var.region
  name        = "${var.rName}-0"
  priority    = 0
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US", "CA"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "${var.rName}-0"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_rule" "test1" {
  region      = var.region
  name        = "${var.rName}-1"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  depends_on = [aws_wafv2_web_acl_rule.test0]

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US", "CA"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "${var.rName}-1"
    sampled_requests_enabled   = false
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resources in"
  type        = string
  nullable    = false
}

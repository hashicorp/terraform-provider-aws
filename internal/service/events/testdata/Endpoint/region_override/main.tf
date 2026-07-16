# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_endpoint" "test" {
  region = var.region

  name = var.rName

  event_bus {
    event_bus_arn = aws_cloudwatch_event_bus.primary.arn
  }
  event_bus {
    event_bus_arn = aws_cloudwatch_event_bus.secondary.arn
  }

  replication_config {
    state = "DISABLED"
  }

  routing_config {
    failover_config {
      primary {
        health_check = aws_route53_health_check.test.arn
      }

      secondary {
        route = var.alt_region
      }
    }
  }
}

data "aws_partition" "current" {}

resource "aws_cloudwatch_event_bus" "primary" {
  region = var.region

  name = var.rName
}

resource "aws_cloudwatch_event_bus" "secondary" {
  provider = "awsalternate"

  name = var.rName
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["events.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_assume.json

  inline_policy {
    name   = var.rName
    policy = data.aws_iam_policy_document.test.json
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "events:PutRule",
      "events:PutTargets",
      "events:DeleteRule",
      "events:RemoveTargets",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:events:*:*:rule/${var.rName}/GlobalEndpointManagedRule-*"]
  }

  statement {
    actions = ["events:PutEvents"]

    resources = [
      aws_cloudwatch_event_bus.primary.arn,
      aws_cloudwatch_event_bus.secondary.arn,
    ]
  }

  statement {
    actions = ["iam:PassRole"]

    resources = ["arn:${data.aws_partition.current.partition}:iam::*:role/${var.rName}"]

    condition {
      test     = "StringLike"
      variable = "iam:PassedToService"

      values = ["events.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_route53_health_check" "test" {
  fqdn             = "example.com"
  type             = "HTTP"
  request_interval = "30"
  disabled         = true
  port             = 80
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

variable "alt_region" {
  description = "Alternate region"
  type        = string
  nullable    = false
}

provider "awsalternate" {
  region = var.alt_region
}
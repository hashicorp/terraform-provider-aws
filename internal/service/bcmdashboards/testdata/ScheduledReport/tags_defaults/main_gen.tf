# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
}

resource "aws_bcmdashboards_scheduled_report" "test" {
  name                                = var.rName
  dashboard_arn                       = aws_bcmdashboards_dashboard.test.arn
  scheduled_report_execution_role_arn = aws_iam_role.test.arn

  schedule_config {
    schedule_expression           = "cron(0 9 1 * ? *)"
    schedule_expression_time_zone = "UTC"
    state                         = "ENABLED"
  }

  depends_on = [aws_iam_role_policy.test]

  tags = var.resource_tags
}

resource "aws_bcmdashboards_dashboard" "test" {
  name = var.rName

  widget {
    title = "example"

    configs {
      query_parameters {
        cost_and_usage {
          granularity = "MONTHLY"
          metrics     = ["UnblendedCost"]

          time_range {
            start_time {
              type  = "ABSOLUTE"
              value = "2025-01-01"
            }
            end_time {
              type  = "ABSOLUTE"
              value = "2025-03-31"
            }
          }
        }
      }

      display_config {
        graph {
          metric      = "UnblendedCost"
          visual_type = "BAR"
        }
      }
    }
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "bcm-dashboards.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["bcm-dashboards:GetDashboard"]
        Resource = "arn:${data.aws_partition.current.partition}:bcm-dashboards::*:dashboard/*"
      },
      {
        Effect = "Allow"
        Action = [
          "ce:GetCostAndUsage",
          "ce:GetDimensionValues",
          "ce:GetTags",
          "ce:GetCostCategories",
          "ce:GetSavingsPlansCoverage",
          "ce:GetReservationUtilization",
          "ce:GetReservationCoverage",
          "ce:GetSavingsPlansUtilization",
        ]
        Resource = "*"
      },
    ]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "provider_tags" {
  type     = map(string)
  nullable = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatchomni_access_grant" "test" {
  name       = var.rName
  domain_id  = aws_cloudwatchomni_domain.test.domain_id
  space_id   = aws_cloudwatchomni_space.test.space_id
  permission = "CUSTOM"

  principal {
    principal_type = "IAM_ROLE"
    principal_id   = aws_iam_role.test.arn
  }

  scoped_action {
    actions = ["cloudwatch:GetRecords"]

    resource {
      resource_type = "DataSet"
      signal_types  = ["LOGS"]

      tags = {
        env = "test"
      }

      row_scope_group {
        row_scope {
          field    = "serviceName"
          operator = "IN"
          values   = ["service-b", "service-a"]
        }

        row_scope {
          field    = "accountId"
          operator = "IN"
          values   = [data.aws_caller_identity.current.account_id]
        }
      }

      row_scope_group {
        row_scope {
          field    = "serviceName"
          operator = "IN"
          values   = ["service-c"]
        }
      }
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_cloudwatchomni_space" "test" {
  name                 = var.rName
  domain_id            = aws_cloudwatchomni_domain.test.domain_id
  data_access_role_arn = aws_iam_role.test.arn
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "cloudwatch.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchAgentServerPolicy"
}

resource "aws_cloudwatchomni_domain" "test" {
  name               = var.rName
  identity_providers = ["IAM"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_redshift_integration" "test" {
  integration_name = var.rName
  source_arn       = aws_dynamodb_table.test.arn
  target_arn       = aws_redshiftserverless_namespace.test.arn

  depends_on = [
    aws_redshiftserverless_workgroup.test,
    aws_redshift_resource_policy.test,
    aws_dynamodb_resource_policy.test,
  ]

  tags = var.resource_tags
}

# testAccIntegrationConfig_source_DynamoDBTable

# The "aws_redshiftserverless_resource_policy" resource doesn't support the following action types.
# Therefore we need to use the "aws_redshift_resource_policy" resource for RedShift-serverless instead.
resource "aws_redshift_resource_policy" "test" {
  resource_arn = aws_redshiftserverless_namespace.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "redshift.amazonaws.com"
        }
        Action   = "redshift:AuthorizeInboundIntegration"
        Resource = aws_redshiftserverless_namespace.test.arn
        Condition = {
          StringEquals = {
            "aws:SourceArn" = aws_dynamodb_table.test.arn
          }
        }
      },
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "redshift:CreateInboundIntegration"
        Resource = aws_redshiftserverless_namespace.test.arn
      }
    ]
  })
}

resource "aws_dynamodb_table" "test" {
  name           = var.rName
  read_capacity  = 1
  write_capacity = 1
  hash_key       = var.rName

  attribute {
    name = var.rName
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}

resource "aws_dynamodb_resource_policy" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "redshift.amazonaws.com"
        }
        Action = [
          "dynamodb:ExportTableToPointInTime",
          "dynamodb:DescribeTable"
        ]
        Resource = aws_dynamodb_table.test.arn
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnEquals = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:redshift:*:${data.aws_caller_identity.current.account_id}:integration:*"
          }
        }
      },
      {
        Effect = "Allow"
        Principal = {
          Service = "redshift.amazonaws.com"
        }
        Action   = "dynamodb:DescribeExport"
        Resource = "${aws_dynamodb_table.test.arn}/export/*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnEquals = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:redshift:*:${data.aws_caller_identity.current.account_id}:integration:*"
          }
        }
      }
    ]
  })
}

# testAccIntegrationConfig_base

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = var.rName
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = var.rName
  base_capacity  = 8

  publicly_accessible = false
  subnet_ids          = aws_subnet.test[*].id
}

# acctest.ConfigVPCWithSubnets(rName, 2)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 2)

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
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

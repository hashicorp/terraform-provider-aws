# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dsql_cluster_policy" "test" {
  count = var.resource_count

  identifier = aws_dsql_cluster.test[count.index].identifier
  region     = var.region

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCurrentAccountConnect"
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.account_id
        }
        Action = [
          "dsql:DbConnect",
        ]
        Resource = aws_dsql_cluster.test[count.index].arn
      }
    ]
  })
}

resource "aws_dsql_cluster" "test" {
  count  = var.resource_count
  region = var.region
}

data "aws_caller_identity" "current" {}

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

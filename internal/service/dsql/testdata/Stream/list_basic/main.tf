# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dsql_stream" "test" {
  count = var.resource_count

  cluster_identifier = aws_dsql_cluster.test.identifier
  format             = "JSON"
  ordering           = "UNORDERED"

  target_definition {
    kinesis {
      role_arn   = aws_iam_role.test.arn
      stream_arn = aws_kinesis_stream.test.arn
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_dsql_cluster" "test" {
}

data "aws_service_principal" "test" {
  service_name = "dsql"
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = data.aws_service_principal.test.name
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["kinesis:DescribeStreamSummary", "kinesis:ListShards", "kinesis:PutRecord", "kinesis:PutRecords"]
        Effect   = "Allow"
        Resource = [aws_kinesis_stream.test.arn]
      },
    ]
  })
}

resource "aws_kinesis_stream" "test" {
  name        = var.rName
  shard_count = 1
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

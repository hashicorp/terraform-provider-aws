# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_s3_table_integration_source" "test" {
  count = var.resource_count

  integration_arn = aws_observabilityadmin_s3_table_integration.test.arn

  data_source {
    name = "test_${count.index}"
    type = "test"
  }
}

# testAccS3TableIntegrationSourceConfig_base

resource "aws_observabilityadmin_s3_table_integration" "test" {
  role_arn = aws_iam_role.test.arn

  encryption {
    sse_algorithm = "AES256"
  }

  depends_on = [aws_iam_role_policy.test]
}

# testAccS3TableIntegrationConfig_base

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "logs.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3tables:CreateTableBucket",
          "s3tables:ListTableBuckets",
          "s3tables:GetTableBucket",
          "s3tables:CreateNamespace",
          "s3tables:GetNamespace",
          "s3tables:ListNamespaces",
          "s3tables:CreateTable",
          "s3tables:GetTable",
          "s3tables:ListTables",
          "s3tables:PutTableData",
          "s3tables:GetTableData",
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

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

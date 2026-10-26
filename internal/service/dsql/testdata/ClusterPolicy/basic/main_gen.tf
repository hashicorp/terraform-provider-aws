# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dsql_cluster_policy" "test" {
  identifier = aws_dsql_cluster.test.identifier

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
          "dsql:DbConnectAdmin",
        ]
        Resource = aws_dsql_cluster.test.arn
      }
    ]
  })
}

data "aws_caller_identity" "current" {}

resource "aws_dsql_cluster" "test" {
}


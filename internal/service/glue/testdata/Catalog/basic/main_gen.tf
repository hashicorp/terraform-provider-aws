# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_glue_catalog" "test" {
  name = var.rName

  catalog_properties {
    data_lake_access_properties {
      catalog_type       = "aws:redshift"
      data_lake_access   = true
      data_transfer_role = aws_iam_role.test.arn
    }
  }

  depends_on = [
    aws_lakeformation_data_lake_settings.test
  ]
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "glue.amazonaws.com",
          "redshift.amazonaws.com",
        ]
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["glue:GetCatalog", "glue:GetDatabase", "kms:Decrypt", "kms:GenerateDataKey"]
      Resource = "*"
    }]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

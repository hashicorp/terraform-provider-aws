# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  region = var.region
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
        Service = ["glue.amazonaws.com", "redshift.amazonaws.com"]
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

resource "aws_glue_catalog" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"

  catalog_properties {
    data_lake_access_properties {
      catalog_type       = "aws:redshift"
      data_lake_access   = true
      data_transfer_role = aws_iam_role.test.arn
    }
  }

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
    aws_iam_role_policy.test,
  ]
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_datazone_environment" "test" {
  name                 = var.rName
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
  ]
}

resource "aws_iam_role" "test" {
  name = var.rName
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = var.rName
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
            "glue:*",
            "lakeformation:*",
            "s3:*",
            "cloudformation:*",
            "athena:*",
            "iam:*",
            "logs:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

data "aws_caller_identity" "test" {}
data "aws_region" "test" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.test.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_iam_session_context.current.issuer_arn,
    aws_iam_role.test.arn,
  ]
}

resource "aws_datazone_domain" "test" {
  name                  = var.rName
  domain_execution_role = aws_iam_role.test.arn

  skip_deletion_check = true
}

resource "aws_security_group" "test" {
  name = var.rName
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  name                = var.rName
  skip_deletion_check = true
}

data "aws_datazone_environment_blueprint" "test" {
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  provisioning_role_arn    = aws_iam_role.test.arn
  manage_access_role_arn   = aws_iam_role.test.arn
  enabled_regions          = [data.aws_region.test.region]

  regional_parameters = {
    (data.aws_region.test.region) = {
      "S3Location" = "s3://${aws_s3_bucket.test.bucket}"
    }
  }
}

resource "aws_datazone_environment_profile" "test" {
  name                             = var.rName
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.47.0"
    }
  }
}

provider "aws" {}

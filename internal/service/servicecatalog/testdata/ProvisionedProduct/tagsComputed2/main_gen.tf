# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_servicecatalog_provisioned_product" "test" {
  name                       = var.rName
  product_id                 = aws_servicecatalog_constraint.test.product_id
  provisioning_artifact_name = var.rName
  path_id                    = data.aws_servicecatalog_launch_paths.test.summaries[0].path_id

  provisioning_parameters {
    key   = "BucketName"
    value = "${var.rName}-dest"
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "${var.rName}.json"

  content = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Parameters = {
      BucketName = {
        Type = "String"
      }
    }

    Resources = {
      MyS3Bucket = {
        Type = "AWS::S3::Bucket"
        Properties = {
          BucketName = { Ref = "BucketName" }
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  description = var.rName
  distributor = "distributör"
  name        = var.rName
  owner       = "ägare"
  type        = "CLOUD_FORMATION_TEMPLATE"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = var.rName
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = var.rName
  description   = var.rName
  provider_name = var.rName
}

resource "aws_servicecatalog_constraint" "test" {
  description  = var.rName
  portfolio_id = aws_servicecatalog_product_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product_portfolio_association.test.product_id
  type         = "RESOURCE_UPDATE"

  parameters = jsonencode({
    Version = "2.0"
    Properties = {
      TagUpdateOnProvisionedProduct = "ALLOWED"
    }
  })
}

resource "aws_servicecatalog_product_portfolio_association" "test" {
  portfolio_id = aws_servicecatalog_principal_portfolio_association.test.portfolio_id
  product_id   = aws_servicecatalog_product.test.id
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_servicecatalog_principal_portfolio_association" "test" {
  portfolio_id  = aws_servicecatalog_portfolio.test.id
  principal_arn = data.aws_iam_session_context.current.issuer_arn # unfortunately, you cannot get launch_path for arbitrary role - only caller
}

data "aws_servicecatalog_launch_paths" "test" {
  product_id = aws_servicecatalog_product_portfolio_association.test.product_id
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}

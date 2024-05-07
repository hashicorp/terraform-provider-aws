# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  default_tags {
    tags = var.provider_tags
  }
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

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "${var.rName}.json"

  content = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}

data "aws_partition" "current" {}


variable "rName" {
  type     = string
  nullable = false
}


variable "provider_tags" {
  type     = map(string)
  nullable = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}

resource "aws_s3_bucket" "test" {
  bucket = var.rName
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
      version = "6.30.0"
    }
  }
}

provider "aws" {}

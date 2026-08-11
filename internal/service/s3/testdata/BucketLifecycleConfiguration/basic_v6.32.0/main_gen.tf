# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_lifecycle_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  rule {
    id     = var.rName
    status = "Enabled"

    expiration {
      days = 365
    }

    filter {
      prefix = "prefix/"
    }
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
      version = "6.32.0"
    }
  }
}

provider "aws" {}

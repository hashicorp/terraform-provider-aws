# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {
  region     = var.minioRegion
  access_key = var.minioAccessKey
  secret_key = var.minioSecretKey

  # Skip AWS-specific validations that will fail against MinIO.
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_requesting_account_id  = true

  # Force path-style URLs (http://minio.local) unless virtual-host style is configured.
  s3_use_path_style = true

  endpoints {
    s3 = var.minioEndpoint
  }
}

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

variable "minioRegion" {
  type     = string
  nullable = false
}

variable "minioAccessKey" {
  type     = string
  nullable = false
}

variable "minioSecretKey" {
  type     = string
  nullable = false
}

variable "minioEndpoint" {
  type     = string
  nullable = false
}
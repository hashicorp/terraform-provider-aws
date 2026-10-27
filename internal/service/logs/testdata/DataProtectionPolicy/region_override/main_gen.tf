# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_data_protection_policy" "test" {
  region = var.region

  log_group_name = aws_cloudwatch_log_group.test.name
  policy_document = jsonencode({
    Name    = "Test"
    Version = "2021-06-01"

    Statement = [
      {
        Sid            = "Audit"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Audit = {
            FindingsDestination = {
              S3 = {
                Bucket = aws_s3_bucket.test.bucket
              }
            }
          }
        }
      },
      {
        Sid            = "Redact"
        DataIdentifier = ["arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress"]
        Operation = {
          Deidentify = {
            MaskConfig = {}
          }
        }
      }
    ]
  })
}

data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  region = var.region

  name = var.rName
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = var.rName
  force_destroy = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

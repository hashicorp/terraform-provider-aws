# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

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
      version = "6.9.0"
    }
  }
}

provider "aws" {}

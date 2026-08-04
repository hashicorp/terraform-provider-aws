# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_policy" "test" {
  count  = var.resource_count
  region = var.region

  bucket = aws_s3_bucket.test[count.index].bucket
  policy = data.aws_iam_policy_document.test[count.index].json
}

data "aws_iam_policy_document" "test" {
  count = var.resource_count

  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      aws_s3_bucket.test[count.index].arn,
      "${aws_s3_bucket.test[count.index].arn}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_s3_bucket" "test" {
  count  = var.resource_count
  region = var.region

  bucket = "${var.rName}-${count.index}"
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

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

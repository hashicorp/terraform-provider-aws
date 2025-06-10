# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_glue_resource_policy" "test" {
  region = var.region

  policy = data.aws_iam_policy_document.glue-example-policy.json
}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = ["glue:CreateTable"]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

data "aws_region" "current" {
  region = var.region
}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

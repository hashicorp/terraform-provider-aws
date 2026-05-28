# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_mlflow_app" "test" {
  name               = var.rName
  artifact_store_uri = "s3://${aws_s3_bucket.test.bucket}/"
  role_arn           = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test_trust.json
}

data "aws_iam_policy_document" "test_trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

data "aws_iam_policy_document" "test_perms" {
  statement {
    effect = "Allow"

    actions = [
      "s3:Get*",
      "s3:Put*",
      "s3:List*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}",
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.bucket}/*"
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "sagemaker:AddTags",
      "sagemaker:CreateModelPackageGroup",
      "sagemaker:CreateModelPackage",
      "sagemaker:UpdateModelPackage",
      "sagemaker:DescribeModelPackageGroup",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_policy" "test" {
  name   = var.rName
  policy = data.aws_iam_policy_document.test_perms.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

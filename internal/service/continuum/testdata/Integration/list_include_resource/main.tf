# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_continuum_integration" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"

  provider_configuration {
    s3 {
      bucket      = aws_s3_bucket.test[count.index].id
      role_arn    = aws_iam_role.test.arn
      external_id = var.rName
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_s3_object.code]
}

resource "aws_s3_bucket" "test" {
  count = var.resource_count

  bucket        = "${var.rName}-${count.index}"
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  count = var.resource_count

  bucket = aws_s3_bucket.test[count.index].id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "code" {
  count = var.resource_count

  bucket     = aws_s3_bucket.test[count.index].id
  key        = "code/app.zip"
  content    = "placeholder"
  depends_on = [aws_s3_bucket_versioning.test]
}

data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["continuum.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = [var.rName]
    }
  }
}

data "aws_iam_policy_document" "access" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = aws_s3_bucket.test[*].arn
  }

  statement {
    actions   = ["s3:GetObject", "s3:GetObjectVersion"]
    resources = flatten([for arn in aws_s3_bucket.test[*].arn : ["${arn}/code/*", "${arn}/context/*"]])
  }

  statement {
    actions   = ["s3:PutObject", "s3:GetObject", "s3:GetObjectVersion"]
    resources = [for arn in aws_s3_bucket.test[*].arn : "${arn}/continuum-patches/*"]
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  path               = "/continuum/integrations/"
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

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

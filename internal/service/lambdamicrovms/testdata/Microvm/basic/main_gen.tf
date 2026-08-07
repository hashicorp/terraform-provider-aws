# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambdamicrovms_microvm" "test" {
  image_identifier = aws_lambdamicrovms_image.test.arn
}

resource "aws_lambdamicrovms_image" "test" {
  name           = var.rName
  base_image_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:aws:microvm-image:al2023-1"
  build_role_arn = aws_iam_role.test.arn

  code_artifact {
    uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = ["s3:GetObject"]
      Effect   = "Allow"
      Resource = "${aws_s3_bucket.test.arn}/*"
    }]
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "code.zip"
  source = "test-fixtures/code.zip"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

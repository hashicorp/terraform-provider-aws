# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambdamicrovms_microvm" "test" {
  image_arn          = aws_lambdamicrovms_image.test.arn
  execution_role_arn = aws_iam_role.test.arn

  logging {
    cloudwatch {
      log_group  = aws_cloudwatch_log_group.test.id
      log_stream = aws_cloudwatch_log_stream.test.id
    }
  }
}

resource "aws_lambdamicrovms_image" "test" {
  name           = var.rName
  base_image_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:aws:microvm-image:al2023-1"
  build_role_arn = aws_iam_role.test.arn

  code_artifact {
    uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = "${var.rName}-s"
  log_group_name = aws_cloudwatch_log_group.test.id
}

resource "aws_cloudwatch_log_group" "test" {
  name = "${var.rName}-g"
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

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.id
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

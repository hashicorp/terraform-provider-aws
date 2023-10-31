# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "example" {
  name               = "tf-ivschat-message-handler-role"
  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Action": ["sts:AssumeRole"],
		"Principal": {"Service": "lambda.amazonaws.com"}
	}]
}
EOF
}

data "archive_file" "message_review_handler" {
  type        = "zip"
  source_file = "${path.module}/index.js"
  output_path = "${path.module}/lambda-handler.zip"
}

resource "aws_lambda_function" "example" {
  filename         = data.archive_file.message_review_handler.output_path
  function_name    = "tf-ivschat-message-handler"
  role             = aws_iam_role.example.arn
  source_code_hash = data.archive_file.message_review_handler.output_base64sha256
  runtime          = "nodejs16.x"
  handler          = "index.handler"
}

resource "aws_lambda_permission" "example" {
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.example.function_name
  principal      = "ivschat.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = "arn:aws:ivschat:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:room/*"
}

resource "aws_s3_bucket" "example" {
  bucket_prefix = "tf-ivschat-logging-bucket-"
  force_destroy = true
}

resource "aws_ivschat_logging_configuration" "example" {
  name = "tf-ivschat-loggingconfiguration"

  lifecycle {
    create_before_destroy = true
  }

  destination_configuration {
    s3 {
      bucket_name = aws_s3_bucket.example.id
    }
  }
}

resource "aws_ivschat_room" "example" {
  name                              = "tf-ivschat-room"
  depends_on                        = [aws_lambda_permission.example]
  logging_configuration_identifiers = [aws_ivschat_logging_configuration.example.arn]

  message_review_handler {
    uri             = aws_lambda_function.example.arn
    fallback_result = "ALLOW"
  }
}

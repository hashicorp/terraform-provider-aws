# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambda_permission" "test" {
  statement_id       = "AllowExecutionFromCloudWatch"
  action             = "lambda:InvokeFunction"
  function_name      = aws_lambda_function.test.function_name
  principal          = "events.amazonaws.com"
  event_source_token = "test-event-source-token"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = var.rName
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

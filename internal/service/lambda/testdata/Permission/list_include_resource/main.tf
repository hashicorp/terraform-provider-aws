# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambda_permission" "test" {
  count = 1

  statement_id  = "${var.rName}-${count.index}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "events.amazonaws.com"
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

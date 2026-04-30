# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lambda_event_source_mapping" "test" {
  count = var.resource_count

  event_source_arn = aws_sqs_queue.test[count.index].arn
  function_name    = aws_lambda_function.test[count.index].arn
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_sqs_queue" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

resource "aws_lambda_function" "test" {
  count = var.resource_count

  filename      = "test-fixtures/lambdatest.zip"
  function_name = "${var.rName}-${count.index}"
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs20.x"
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

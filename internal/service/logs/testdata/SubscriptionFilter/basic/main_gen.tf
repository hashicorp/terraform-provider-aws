# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_subscription_filter" "test" {
  destination_arn = aws_kinesis_stream.test.arn
  filter_pattern  = "logtype test"
  log_group_name  = aws_cloudwatch_log_group.test.name
  name            = "${var.rName}-filter"
  role_arn        = aws_iam_role.test.arn
}

data "aws_region" "current" {
}

resource "aws_cloudwatch_log_group" "test" {
  name              = "${var.rName}-group"
  retention_in_days = 1
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
        "Service": "logs.${data.aws_region.current.region}.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "kinesis:PutRecord",
      "Resource": "${aws_kinesis_stream.test.arn}"
    },
    {
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.test.arn}"
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "test" {
  name        = var.rName
  shard_count = 1
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

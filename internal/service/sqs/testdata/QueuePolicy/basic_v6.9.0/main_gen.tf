# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": "sqs:*",
    "Resource": "${aws_sqs_queue.test.arn}",
    "Condition": {
      "ArnEquals": {
        "aws:SourceArn": "${aws_sqs_queue.test.arn}"
      }
    }
  }]
}
POLICY
}

resource "aws_sqs_queue" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.9.0"
    }
  }
}

provider "aws" {}

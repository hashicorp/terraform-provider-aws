# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sqs_queue_policy" "test" {
  count  = var.resource_count
  region = var.region

  queue_url = aws_sqs_queue.test[count.index].id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": "sqs:*",
    "Resource": "${aws_sqs_queue.test[count.index].arn}",
    "Condition": {
      "ArnEquals": {
        "aws:SourceArn": "${aws_sqs_queue.test[count.index].arn}"
      }
    }
  }]
}
POLICY
}

resource "aws_sqs_queue" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

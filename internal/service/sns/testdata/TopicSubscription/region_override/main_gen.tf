# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sns_topic" "test" {
  region = var.region

  name = var.rName
}

resource "aws_sqs_queue" "test" {
  region = var.region

  name = var.rName

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  region = var.region

  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

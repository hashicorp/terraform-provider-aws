# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sns_topic" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"
}

resource "aws_sqs_queue" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  count  = var.resource_count
  region = var.region

  topic_arn = aws_sns_topic.test[count.index].arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test[count.index].arn
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

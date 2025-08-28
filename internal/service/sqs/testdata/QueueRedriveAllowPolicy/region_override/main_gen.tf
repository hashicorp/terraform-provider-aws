# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sqs_queue_redrive_allow_policy" "test" {
  region = var.region

  queue_url = aws_sqs_queue.test.id
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.test_src.arn]
  })
}

resource "aws_sqs_queue" "test" {
  region = var.region

  name = var.rName
}

resource "aws_sqs_queue" "test_src" {
  region = var.region

  name = "${var.rName}_src"
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test.arn
    maxReceiveCount     = 4
  })
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

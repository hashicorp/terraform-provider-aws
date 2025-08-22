# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sqs_queue_redrive_policy" "test" {
  queue_url = aws_sqs_queue.test.id
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.test_ddl.arn
    maxReceiveCount     = 4
  })
}

resource "aws_sqs_queue" "test" {
  name = var.rName
}

resource "aws_sqs_queue" "test_ddl" {
  name = "${var.rName}_ddl"
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.test.arn]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

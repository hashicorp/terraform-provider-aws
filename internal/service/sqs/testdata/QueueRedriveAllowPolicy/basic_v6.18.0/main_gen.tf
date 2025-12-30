# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_sqs_queue_redrive_allow_policy" "test" {
  queue_url = aws_sqs_queue.test.id
  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.test_src.arn]
  })
}

resource "aws_sqs_queue" "test" {
  name = var.rName
}

resource "aws_sqs_queue" "test_src" {
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
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.18.0"
    }
  }
}

provider "aws" {}

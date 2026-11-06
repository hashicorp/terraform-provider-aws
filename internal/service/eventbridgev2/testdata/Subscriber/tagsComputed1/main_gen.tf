# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_eventbridgev2_subscriber" "test" {
  name          = var.rName
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.test.arn
    target_arn = aws_sqs_queue.test.arn
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_eventbridgev2_event_bus" "test" {
  name = var.rName
}

resource "aws_sqs_queue" "test" {
  name = var.rName
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "events.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "sqs:SendMessage"
      Effect   = "Allow"
      Resource = aws_sqs_queue.test.arn
    }]
  })
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

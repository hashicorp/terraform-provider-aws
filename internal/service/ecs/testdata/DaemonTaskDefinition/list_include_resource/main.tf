# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_daemon_task_definition" "test" {
  count = var.resource_count

  family             = "${var.rName}-${count.index}"
  execution_role_arn = aws_iam_role.test.arn

  container_definition {
    name      = "test"
    image     = "nginx:latest"
    essential = true
    memory    = 128
  }
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "The number of resources to create"
  type        = number
  nullable    = false
}


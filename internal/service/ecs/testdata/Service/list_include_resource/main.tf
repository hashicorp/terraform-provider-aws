# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_service" "test" {
  count = var.resource_count

  name            = "${var.rName}-${count.index}"
  cluster         = aws_ecs_cluster.test.arn
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  tags = var.resource_tags
}

resource "aws_ecs_cluster" "test" {
  name = var.rName
}

resource "aws_ecs_task_definition" "test" {
  family = var.rName

  container_definitions = jsonencode([{
    name      = "test"
    image     = "public.ecr.aws/docker/library/busybox:latest"
    cpu       = 128
    memory    = 128
    essential = true
  }])
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

variable "resource_tags" {
  description = "Tags to set on resource"
  type        = map(string)
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_task_definition" "test" {
  count = var.resource_count

  family = "${var.rName}-${count.index}"

  container_definitions = jsonencode([{
    name      = "test"
    image     = "nginx:latest"
    cpu       = 10
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

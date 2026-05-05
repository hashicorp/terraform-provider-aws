# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_service" "test" {
  name            = var.rName
  cluster         = aws_ecs_cluster.test.arn
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}

resource "aws_ecs_cluster" "test" {
  name = var.rName
}

resource "aws_ecs_task_definition" "test" {
  family = var.rName

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

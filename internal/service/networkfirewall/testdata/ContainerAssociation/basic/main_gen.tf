# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_cluster" "test" {
  name = var.rName
}

resource "aws_networkfirewall_container_association" "test" {
  container_association_name = var.rName
  type                       = "ECS"

  container_monitoring_configuration {
    cluster_arn = aws_ecs_cluster.test.arn
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

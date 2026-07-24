# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_cluster" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
}

resource "aws_networkfirewall_container_association" "test" {
  count = var.resource_count

  container_association_name = "${var.rName}-${count.index}"
  type                       = "ECS"

  container_monitoring_configurations {
    cluster_arn = aws_ecs_cluster.test[count.index].arn
  }
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

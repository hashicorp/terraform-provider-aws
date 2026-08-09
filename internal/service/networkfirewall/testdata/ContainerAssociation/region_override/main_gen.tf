# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_cluster" "test" {
  region = var.region

  name = var.rName
}

resource "aws_networkfirewall_container_association" "test" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

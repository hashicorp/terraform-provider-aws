# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_appautoscaling_target" "test" {
  count  = var.resource_count
  region = var.region

  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test[count.index].name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15
}

resource "aws_dynamodb_table" "test" {
  count  = var.resource_count
  region = var.region

  name           = "${var.rName}-${count.index}"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "TestKey"

  attribute {
    name = "TestKey"
    type = "S"
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

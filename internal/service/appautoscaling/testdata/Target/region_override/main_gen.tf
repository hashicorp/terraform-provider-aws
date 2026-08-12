# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_appautoscaling_target" "test" {
  region = var.region

  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15
}

resource "aws_dynamodb_table" "test" {
  region = var.region

  name           = var.rName
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

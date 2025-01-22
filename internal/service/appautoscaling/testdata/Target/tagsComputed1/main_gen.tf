# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 2
  max_capacity       = 15

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_dynamodb_table" "test" {
  name           = var.rName
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "TestKey"

  attribute {
    name = "TestKey"
    type = "S"
  }
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

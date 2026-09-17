# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dynamodb_global_secondary_index" "test" {
  region = var.region

  table_name = aws_dynamodb_table.test.name
  index_name = var.rName

  projection {
    projection_type = "ALL"
  }

  provisioned_throughput {
    read_capacity_units  = 1
    write_capacity_units = 1
  }

  key_schema {
    attribute_name = var.rName
    attribute_type = "S"
    key_type       = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  region = var.region

  name           = var.rName
  hash_key       = var.rName
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = var.rName
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

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dynamodb_table_item" "test" {
  count  = var.resource_count
  region = var.region

  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
{
  "TestTableHashKey": {"S": "${var.rName}-${count.index}"},
  "one": {"N": "11111"}
}
ITEM
}

resource "aws_dynamodb_table" "test" {
  region = var.region

  name           = var.rName
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
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

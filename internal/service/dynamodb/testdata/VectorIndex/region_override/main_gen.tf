# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dynamodb_vector_index" "test" {
  region = var.region

  table_name = aws_dynamodb_table.test.name
  index_name = var.rName

  dimensions        = 2
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = var.rName
    attribute_type = "S"
    type           = "HASH"
  }
}

resource "aws_dynamodb_table" "test" {
  region = var.region

  name           = var.rName
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = var.rName

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

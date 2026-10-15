# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.test.arn

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_dynamodb_table" "test" {
  region = var.secondary_region

  name             = var.rName
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  tags = {
    # Should not show up on `aws_dynamodb_table_replica`
    Name = var.rName
  }

  lifecycle {
    ignore_changes = [replica]
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

variable "secondary_region" {
  description = "Secondary region"
  type        = string
  nullable    = false
}

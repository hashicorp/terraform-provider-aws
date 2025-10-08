# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "awsalternate" {
  region = var.alt_region
}

provider "null" {}

resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.test.arn

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "aws_dynamodb_table" "test" {
  provider         = "awsalternate"
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

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}

variable "alt_region" {
  description = "Region for provider awsalternate"
  type        = string
  nullable    = false
}

# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "awsalternate" {
  region = var.alt_region
}

resource "aws_dynamodb_table_replica" "test" {
  global_table_arn = aws_dynamodb_table.test.arn

  tags = var.resource_tags
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

variable "alt_region" {
  description = "Region for provider awsalternate"
  type        = string
  nullable    = false
}

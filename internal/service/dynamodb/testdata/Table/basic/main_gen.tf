# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = var.rName
  read_capacity  = 1
  write_capacity = 1

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

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_dynamodb_global_secondary_index" "test" {
  provider = aws

  include_resource = true

  config {
    table_name = aws_dynamodb_table.test.name
  }
}

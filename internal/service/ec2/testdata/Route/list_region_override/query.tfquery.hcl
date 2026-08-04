# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_route" "test" {
  provider = aws

  config {
    region         = var.region
    route_table_id = aws_route_table.test.id
  }
}

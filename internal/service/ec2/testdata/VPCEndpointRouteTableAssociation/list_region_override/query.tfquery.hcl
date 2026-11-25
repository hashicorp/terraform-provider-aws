# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_vpc_endpoint_route_table_association" "test" {
  provider = aws

  config {
    region = var.region
  }
}

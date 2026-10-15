# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

list "aws_ec2_transit_gateway_route_table_propagation" "test" {
  provider = aws

  config {
    region                         = var.region
    transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
  }
}

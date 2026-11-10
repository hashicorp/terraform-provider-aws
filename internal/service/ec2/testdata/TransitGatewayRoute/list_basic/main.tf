# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_route" "test" {
  count = var.resource_count

  blackhole                      = true
  destination_cidr_block         = "10.${count.index}.0.0/16"
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway" "test" {
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

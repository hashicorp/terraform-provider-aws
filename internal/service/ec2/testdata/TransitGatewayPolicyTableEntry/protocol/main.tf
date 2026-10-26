# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 400
  target_route_table_id           = aws_ec2_transit_gateway_route_table.test.id

  policy_rule {
    source_cidr_block      = "10.0.1.0/24"
    destination_cidr_block = "10.0.2.0/24"
    protocol               = var.protocol
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_policy_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = var.rName
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "protocol" {
  type     = string
  nullable = false
}
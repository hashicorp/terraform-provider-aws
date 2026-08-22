# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_policy_table_entry" "test" {
  region = var.region

  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.test.id
  policy_rule_number              = 100
  target_route_table_id           = aws_ec2_transit_gateway_route_table.test.id
}

resource "aws_ec2_transit_gateway" "test" {
  region = var.region

  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_policy_table" "test" {
  region = var.region

  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = var.rName
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

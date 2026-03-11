# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_metering_policy" "test" {
  count = var.resource_count

  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway" "test" {}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

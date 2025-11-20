# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_vpn_concentrator" "test" {
  type               = "ipsec.1"
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = var.resource_tags
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

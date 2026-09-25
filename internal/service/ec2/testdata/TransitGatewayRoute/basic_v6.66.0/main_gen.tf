# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_route" "test" {
  blackhole                      = true
  destination_cidr_block         = "10.1.0.0/16"
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}

resource "aws_ec2_transit_gateway" "test" {
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.66.0"
    }
  }
}

provider "aws" {}

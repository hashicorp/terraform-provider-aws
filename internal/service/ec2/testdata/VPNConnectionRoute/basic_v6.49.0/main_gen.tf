# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = 65000
  ip_address = "182.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = var.rName
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    Name = var.rName
  }
}

resource "aws_vpn_connection_route" "test" {
  destination_cidr_block = "172.168.10.0/24"
  vpn_connection_id      = aws_vpn_connection.test.id
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
      version = "6.49.0"
    }
  }
}

provider "aws" {}

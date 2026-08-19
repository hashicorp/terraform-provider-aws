# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = var.rName
  }
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = var.rName
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = var.rName

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["443"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_devopsagent_private_connection" "test" {

  name                      = var.rName
  mode                      = "SELF_MANAGED"
  resource_configuration_id = aws_vpclattice_resource_configuration.test.id

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}

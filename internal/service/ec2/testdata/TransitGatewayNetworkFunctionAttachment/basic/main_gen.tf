# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_transit_gateway_network_function_attachment" "test" {
  availability_zone_ids = [aws_subnet.test[0].availability_zone_id]
  service_name          = aws_vpc_endpoint_service.test.service_name
  transit_gateway_id    = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = var.rName
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = var.rName
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = var.rName
  subnets            = aws_subnet.test[*].id
}

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 1)

resource "aws_subnet" "test" {
  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

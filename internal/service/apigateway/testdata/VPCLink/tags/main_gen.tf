# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_api_gateway_vpc_link" "test" {
  name        = var.rName
  target_arns = [aws_lb.test.arn]

  tags = var.resource_tags
}

resource "aws_lb" "test" {
  name               = var.rName
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id
}

// acctest.ConfigVPCWithSubnets(rName, 1)
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

resource "aws_subnet" "test" {
  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

// acctest.ConfigAvailableAZsNoOptInDefaultExclude()
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

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

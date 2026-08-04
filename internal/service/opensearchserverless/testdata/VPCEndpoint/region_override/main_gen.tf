# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_vpc_endpoint" "test" {
  region = var.region

  name               = var.rName
  subnet_ids         = [aws_subnet.test[0].id]
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test[0].id]
}

resource "aws_subnet" "test" {
  region = var.region

  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

data "aws_availability_zones" "available" {
  region = var.region

  exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = var.rName
  }
}

resource "aws_security_group" "test" {
  region = var.region

  count  = 2
  name   = "${var.rName}-${count.index}"
  vpc_id = aws_vpc.test.id

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

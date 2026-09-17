# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc_endpoint" "expected" {
  count = var.resource_count

  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name     = "${var.rName}-expected-${count.index}"
    expected = true
  }
}

resource "aws_vpc_endpoint" "not_expected" {
  count = var.resource_count

  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.region}.s3"

  tags = {
    Name     = "${var.rName}-not-expected-${count.index}"
    expected = false
  }
}


resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.rName
  }
}

data "aws_region" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

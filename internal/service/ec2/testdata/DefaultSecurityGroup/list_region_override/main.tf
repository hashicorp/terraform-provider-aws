# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  count = var.resource_count

  region     = var.region
  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

resource "aws_default_security_group" "test" {
  count = var.resource_count

  region = var.region
  vpc_id = aws_vpc.test[count.index].id
}

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

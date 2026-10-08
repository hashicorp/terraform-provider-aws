# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_vpc" "test" {
  count = var.resource_count

  cidr_block = "10.${count.index}.0.0/16"

  tags = {
    Name = "${var.rName}-${count.index}"
  }
}

resource "aws_default_security_group" "test" {
  count = var.resource_count

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

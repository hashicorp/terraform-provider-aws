# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lb_target_group" "test" {
  count  = var.resource_count
  region = var.region

  name     = "${var.rName}-${count.index}"
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  region = var.region

  cidr_block = "10.0.0.0/16"
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

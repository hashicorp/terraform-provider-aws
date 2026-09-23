# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ram_resource_association" "test" {
  region = var.region

  resource_arn       = aws_ec2_managed_prefix_list.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_share" "test" {
  region = var.region

  name = var.rName
}

resource "aws_ec2_managed_prefix_list" "test" {
  region = var.region

  name           = var.rName
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr        = "10.0.0.0/8"
    description = "Test entry"
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

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ec2_managed_prefix_list.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_resource_share" "test" {
  name = var.rName
}

resource "aws_ec2_managed_prefix_list" "test" {
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
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.63.0"
    }
  }
}

provider "aws" {}

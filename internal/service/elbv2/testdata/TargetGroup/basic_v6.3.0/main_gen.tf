# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lb_target_group" "test" {
  name     = var.rName
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
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
      version = "6.3.0"
    }
  }
}

provider "aws" {}

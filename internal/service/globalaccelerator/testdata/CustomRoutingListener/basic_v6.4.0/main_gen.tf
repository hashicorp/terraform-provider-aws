# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = var.rName
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.arn

  port_range {
    from_port = 443
    to_port   = 443
  }

  port_range {
    from_port = 10000
    to_port   = 30000
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
      version = "6.4.0"
    }
  }
}

provider "aws" {}

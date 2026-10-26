# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_event_connection" "test" {
  name               = var.rName
  authorization_type = "BASIC"

  auth_parameters {
    basic {
      username = "${var.rName}-user"
      password = "${var.rName}-pass"
    }
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
      version = "6.53.0"
    }
  }
}

provider "aws" {}

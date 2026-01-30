# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_devicefarm_network_profile" "test" {
  name        = var.rName
  project_arn = aws_devicefarm_project.test.arn
}

# testAccProjectConfig_basic

resource "aws_devicefarm_project" "test" {
  name = var.rName
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
      version = "5.100.0"
    }
  }
}

provider "aws" {}

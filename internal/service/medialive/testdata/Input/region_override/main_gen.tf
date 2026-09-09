# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_medialive_input" "test" {
  region = var.region

  name                  = var.rName
  input_security_groups = [aws_medialive_input_security_group.test.id]
  type                  = "UDP_PUSH"
}

# testAccInputBaseConfig

resource "aws_medialive_input_security_group" "test" {
  region = var.region

  whitelist_rules {
    cidr = "10.0.0.8/32"
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

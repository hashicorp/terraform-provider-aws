# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_mailmanager_rule_set" "test" {
  region = var.region

  name = var.rName

  rule {
    name = "example"

    action {
      add_header {
        header_name  = "X-Example"
        header_value = "example"
      }
    }
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

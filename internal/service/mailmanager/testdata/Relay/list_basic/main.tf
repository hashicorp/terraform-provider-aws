# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_mailmanager_relay" "test" {
  count = var.resource_count

  name        = "${var.rName}-${count.index}"
  server_name = "smtp.example.com"
  server_port = 25

  authentication {
    no_authentication {}
  }
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

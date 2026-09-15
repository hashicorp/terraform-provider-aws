# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_key_pair" "test" {
  key_name   = var.rName
  public_key = var.public_key
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
variable "public_key" {
  type     = string
  nullable = false
}


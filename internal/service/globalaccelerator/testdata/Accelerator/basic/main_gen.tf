# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_globalaccelerator_accelerator" "test" {
  name            = var.rName
  ip_address_type = "IPV4"
  enabled         = false

  tags = var.resource_tags
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a type to allow for `null` value
  default = null
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

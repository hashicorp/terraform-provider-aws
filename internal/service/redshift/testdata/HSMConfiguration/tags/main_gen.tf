# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_redshift_hsm_configuration" "test" {
  description                   = var.rName
  hsm_configuration_identifier  = var.rName
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = var.rName
  hsm_server_public_certificate = var.rName

  tags = var.resource_tags
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

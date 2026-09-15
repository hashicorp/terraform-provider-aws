# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_drs_launch_configuration_template" "test" {
  copy_private_ip             = true
  copy_tags                   = false
  launch_disposition          = "STARTED"
  launch_into_source_instance = false
  post_launch_enabled         = false
  recovery_mode               = "OPTIMAL"

  licensing {
    os_byol = true
  }

  target_instance_type_right_sizing_method = "IN_AWS"

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

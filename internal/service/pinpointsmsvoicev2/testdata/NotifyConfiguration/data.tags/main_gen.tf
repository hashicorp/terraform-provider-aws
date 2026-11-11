# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

# tflint-ignore: terraform_unused_declarations
data "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  id = aws_pinpointsmsvoicev2_notify_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_notify_configuration" "test" {
  display_name     = var.rName
  use_case         = "CODE_VERIFICATION"
  enabled_channels = ["SMS"]

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

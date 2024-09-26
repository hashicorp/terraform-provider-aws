# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"
  name           = var.rName

  tags = var.resource_tags
}

resource "aws_appconfig_application" "test" {
  name = var.rName
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

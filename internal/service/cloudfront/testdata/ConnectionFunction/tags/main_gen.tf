# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudfront_connection_function" "test" {
  name                     = var.rName
  connection_function_code = "function handler(event) { return event.request; }"

  connection_function_config {
    comment = "Test connection function"
    runtime = "cloudfront-js-2.0"
  }

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

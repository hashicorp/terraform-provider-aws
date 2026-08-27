# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_application_status_check" "test" {
  protocol = "http"
  port     = 80

  tags = var.resource_tags
}

variable "resource_tags" {
  description = "Tags to set on the resource"
  type        = map(string)
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = var.resource_tags
}

resource "aws_networkmanager_global_network" "test" {}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

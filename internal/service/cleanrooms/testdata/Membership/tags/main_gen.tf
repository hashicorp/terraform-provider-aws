# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_cleanrooms_collaboration" "test" {
  name                     = var.rName
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  creator_display_name     = "Creator"
  description              = var.rName
  query_log_status         = "DISABLED"
}

resource "aws_cleanrooms_membership" "test" {
  collaboration_id = aws_cleanrooms_collaboration.test.id
  query_log_status = "DISABLED"

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

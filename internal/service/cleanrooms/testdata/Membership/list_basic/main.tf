# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cleanrooms_membership" "test" {
  count = var.resource_count

  collaboration_id = aws_cleanrooms_collaboration.test[count.index].id
  query_log_status = "DISABLED"
}

resource "aws_cleanrooms_collaboration" "test" {
  count = var.resource_count

  name                     = "${var.rName}-${count.index}"
  creator_display_name     = "Creator"
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  description              = "Test membership ${count.index}"
  query_log_status         = "DISABLED"
  analytics_engine         = "SPARK"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

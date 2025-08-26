# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

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

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

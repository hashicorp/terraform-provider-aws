# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cleanrooms_membership" "test" {
  region = var.region

  collaboration_id = aws_cleanrooms_collaboration.test.id
  query_log_status = "DISABLED"
}

resource "aws_cleanrooms_collaboration" "test" {
  region = var.region

  name                     = var.rName
  creator_member_abilities = ["CAN_QUERY", "CAN_RECEIVE_RESULTS"]
  creator_display_name     = "Creator"
  description              = var.rName
  query_log_status         = "DISABLED"
  analytics_engine         = "SPARK"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

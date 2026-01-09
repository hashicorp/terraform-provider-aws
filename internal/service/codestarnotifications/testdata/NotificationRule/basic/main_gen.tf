# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = var.rName
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}

# testAccNotificationRuleConfig_base

resource "aws_codecommit_repository" "test" {
  repository_name = var.rName
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

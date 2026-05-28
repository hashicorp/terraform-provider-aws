# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_event_destination" "test" {
  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = var.rName

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test.arn
  }
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = var.rName
}

resource "aws_sns_topic" "test" {
  name = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

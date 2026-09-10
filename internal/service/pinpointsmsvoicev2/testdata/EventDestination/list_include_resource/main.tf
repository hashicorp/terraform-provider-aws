# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_event_destination" "test" {
  count = var.resource_count

  configuration_set_name = aws_pinpointsmsvoicev2_configuration_set.test.name
  event_destination_name = "${var.rName}-${count.index}"

  matching_event_types = ["TEXT_DELIVERED"]

  sns_destination {
    topic_arn = aws_sns_topic.test[count.index].arn
  }
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = var.rName
}

resource "aws_sns_topic" "test" {
  count = var.resource_count

  name = "${var.rName}-${count.index}"
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

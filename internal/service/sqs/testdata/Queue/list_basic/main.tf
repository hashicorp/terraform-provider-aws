# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_sqs_queue" "test" {
  count                     = 2
  name                      = "${var.rName}-${count.index}"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
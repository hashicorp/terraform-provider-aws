# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_protect_configuration_association" "test" {
  configuration_set_name   = aws_pinpointsmsvoicev2_configuration_set.test.name
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = var.rName
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

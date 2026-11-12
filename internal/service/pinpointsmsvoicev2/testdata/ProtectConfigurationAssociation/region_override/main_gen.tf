# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_protect_configuration_association" "test" {
  region = var.region

  configuration_set_name   = aws_pinpointsmsvoicev2_configuration_set.test.name
  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  region = var.region

  name = var.rName
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
  region = var.region

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

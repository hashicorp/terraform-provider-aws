# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_pinpointsmsvoicev2_account_default_protect_configuration" "test" {
  region = var.region

  protect_configuration_id = aws_pinpointsmsvoicev2_protect_configuration.test.id
}

resource "aws_pinpointsmsvoicev2_protect_configuration" "test" {
  region = var.region

}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

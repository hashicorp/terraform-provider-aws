# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_application_status_check" "test" {
  region = var.region

  protocol = "http"
  port     = 80
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_application_status_check_association" "test" {
  region = var.region

  application_status_check_id = aws_ec2_application_status_check.test.id
  target_tag_key              = "Name"
  target_tag_value            = var.rName
}

resource "aws_ec2_application_status_check" "test" {
  region = var.region

  protocol = "http"
  port     = 80
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

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_application_status_check_association" "test" {
  application_status_check_id = aws_ec2_application_status_check.test.id
  target_tag_key              = "Name"
  target_tag_value            = var.rName
}

resource "aws_ec2_application_status_check" "test" {
  protocol = "http"
  port     = 80
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

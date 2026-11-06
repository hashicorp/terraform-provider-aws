# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_application_status_check" "test" {
  count = var.resource_count

  region = var.region

  protocol = "http"
  port     = 80
}

resource "aws_ec2_application_status_check_association" "test" {
  count = var.resource_count

  region = var.region

  application_status_check_id = aws_ec2_application_status_check.test[count.index].id
  target_tag_key              = "Environment"
  target_tag_value            = "production"
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "AWS region"
  type        = string
  nullable    = false
}

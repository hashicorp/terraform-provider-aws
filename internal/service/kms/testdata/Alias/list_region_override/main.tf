# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_kms_key" "test" {
  region = var.region
  count  = var.resource_count

  description             = "${var.rName}-${count.index}"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  region = var.region
  count  = var.resource_count

  name          = "alias/${var.rName}-${count.index}"
  target_key_id = aws_kms_key.test[count.index].key_id
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_directoryservicedata_user" "test" {
  count = var.resource_count

  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = "${var.rName}-${count.index}"
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

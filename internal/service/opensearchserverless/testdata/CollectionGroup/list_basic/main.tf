# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_collection_group" "test" {
  count = var.resource_count

  name             = "${substr(var.rName, 0, 30)}-${count.index}"
  standby_replicas = "ENABLED"
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

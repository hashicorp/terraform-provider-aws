# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_collection_group" "test" {
  name             = var.rName
  standby_replicas = "ENABLED"

  capacity_limits {
    max_indexing_capacity_in_ocu = 1
    max_search_capacity_in_ocu   = 1
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

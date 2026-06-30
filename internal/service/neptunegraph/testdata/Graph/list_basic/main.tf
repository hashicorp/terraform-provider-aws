# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_neptunegraph_graph" "test" {
  graph_name          = var.rName
  provisioned_memory  = 16
  deletion_protection = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

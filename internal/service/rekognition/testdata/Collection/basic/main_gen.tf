# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_rekognition_collection" "test" {
  collection_id = var.rName
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

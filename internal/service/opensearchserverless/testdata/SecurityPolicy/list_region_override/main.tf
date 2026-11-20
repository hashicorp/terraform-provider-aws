# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_security_policy" "test" {
  count  = var.resource_count
  region = var.region

  name = "${substr(var.rName, 0, 30)}-${count.index}"
  type = "encryption"
  policy = jsonencode({
    Rules = [
      {
        Resource = [
          "collection/${substr(var.rName, 0, 30)}-${count.index}"
        ],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = true
  })
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

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_security_policy" "test" {
  count = var.resource_count

  name = "tf-${count.index}-${substr(var.rName, 0, 20)}"
  type = "encryption"
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/tf-${count.index}-${substr(var.rName, 0, 20)}"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
  count = var.resource_count

  name = "tf-${count.index}-${substr(var.rName, 0, 20)}"

  depends_on = [aws_opensearchserverless_security_policy.test]
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

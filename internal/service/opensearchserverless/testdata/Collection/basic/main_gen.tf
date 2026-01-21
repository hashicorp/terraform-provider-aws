# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_security_policy" "test" {
  name = var.rName
  type = "encryption"
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/${var.rName}"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
  name = var.rName

  depends_on = [aws_opensearchserverless_security_policy.test]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

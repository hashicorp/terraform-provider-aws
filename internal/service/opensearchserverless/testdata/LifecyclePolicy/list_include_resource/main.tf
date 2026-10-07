# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_lifecycle_policy" "test" {
  count = var.resource_count

  name = "${substr(var.rName, 0, 30)}-${count.index}"
  type = "retention"
  policy = jsonencode({
    "Rules" : [
      {
        "ResourceType" : "index",
        "Resource" : ["index/${var.rName}-${count.index}/*"],
        "MinIndexRetention" : "81d"
      }
    ]
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

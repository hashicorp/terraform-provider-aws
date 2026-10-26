# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_access_policy" "test" {
  count = var.resource_count

  name = "${substr(var.rName, 0, 30)}-${count.index}"
  type = "data"
  policy = jsonencode([
    {
      Rules : [
        {
          ResourceType : "index",
          Resource : ["index/books/*"],
          Permission : [
            "aoss:CreateIndex",
            "aoss:ReadDocument",
            "aoss:UpdateIndex",
            "aoss:DeleteIndex",
            "aoss:WriteDocument"
          ]
        }
      ],
      Principal : [
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin"
      ]
    }
  ])
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

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

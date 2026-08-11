# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
  region = var.region

  name = var.rName
  type = "data"
  policy = jsonencode([
    {
      Rules : [
        {
          ResourceType : "index",
          Resource : [
            "index/books/*"
          ],
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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_lifecycle_policy" "test" {
  region = var.region

  name = var.rName
  type = "retention"
  policy = jsonencode({
    Rules : [
      {
        ResourceType : "index",
        Resource : ["index/${var.rName}/*"],
        MinIndexRetention : "81d"
      },
      {
        ResourceType : "index",
        Resource : ["index/sales/${var.rName}*"],
        NoMinIndexRetention : true
      }
    ]
  })
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

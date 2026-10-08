# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_opensearchserverless_lifecycle_policy" "test" {
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
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.39.0"
    }
  }
}

provider "aws" {}

# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
  name = var.rName
}

resource "aws_sns_topic_data_protection_policy" "test" {
  arn = aws_sns_topic.test.arn
  policy = jsonencode(
    {
      "Description" = "Default data protection policy"
      "Name"        = "__default_data_protection_policy"
      "Statement" = [
        {
          "DataDirection" = "Inbound"
          "DataIdentifier" = [
            "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
          ]
          "Operation" = {
            "Deny" = {}
          }
          "Principal" = [
            "*",
          ]
          "Sid" = var.rName
        },
      ]
      "Version" = "2021-06-01"
    }
  )
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

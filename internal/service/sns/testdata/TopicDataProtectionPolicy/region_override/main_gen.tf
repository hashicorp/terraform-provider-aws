# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
  region = var.region

  name = var.rName
}

resource "aws_sns_topic_data_protection_policy" "test" {
  region = var.region

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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

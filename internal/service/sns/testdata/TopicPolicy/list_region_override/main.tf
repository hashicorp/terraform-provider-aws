# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sns_topic" "test" {
  count  = var.resource_count
  region = var.region

  name = "${var.rName}-${count.index}"
}

resource "aws_sns_topic_policy" "test" {
  count  = var.resource_count
  region = var.region

  arn = aws_sns_topic.test[count.index].arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "${var.rName}-${count.index}"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action = [
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
      ]
      Resource = aws_sns_topic.test[count.index].arn
    }]
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

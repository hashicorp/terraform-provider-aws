# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_eventbridgev2_resource_policy" "test" {
  count = var.resource_count

  region = var.region

  resource_arn = aws_eventbridgev2_event_bus.test[count.index].arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowAccount"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action   = "events:PutEvents"
      Resource = aws_eventbridgev2_event_bus.test[count.index].arn
    }]
  })
}

resource "aws_eventbridgev2_event_bus" "test" {
  count = var.resource_count

  region = var.region

  name = "${var.rName}-${count.index}"
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

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

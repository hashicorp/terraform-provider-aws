# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_delivery_destination_policy" "test" {
  delivery_destination_name   = aws_cloudwatch_log_delivery_destination.test.name
  delivery_destination_policy = <<EOF
{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:delivery-source:*",
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:delivery:*",
        "arn:${data.aws_partition.current.partition}:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:delivery-destination:*"
      ],
      "Principal":{
        "AWS":"arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action":"logs:CreateDelivery"
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = var.rName
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name = var.rName

  delivery_destination_configuration {
    destination_resource_arn = aws_cloudwatch_log_group.test.arn
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

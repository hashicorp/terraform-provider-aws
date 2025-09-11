# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sns_topic" "test" {
  name = var.rName
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"${var.rName}",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

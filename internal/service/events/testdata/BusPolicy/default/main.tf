# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0


resource "aws_cloudwatch_event_bus_policy" "test" {
  policy = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    sid    = "test-resource-policy"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutEvents",
      "events:PutRule"
    ]
    resources = [
      "*",
    ]
  }
}

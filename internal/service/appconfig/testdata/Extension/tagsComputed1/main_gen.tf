# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_appconfig_extension" "test" {
  name = var.rName
  action_point {
    point = "ON_DEPLOYMENT_COMPLETE"
    action {
      name     = "test"
      role_arn = aws_iam_role.test.arn
      uri      = aws_sns_topic.test.arn
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
  }
}

resource "aws_sns_topic" "test" {
  name =var.rName
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["appconfig.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = var.rName
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

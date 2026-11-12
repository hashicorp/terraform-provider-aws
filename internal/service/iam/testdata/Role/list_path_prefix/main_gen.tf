# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

variable "rName" {
  type = string
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test_match" {
  count              = 2
  name               = "${var.rName}-${count.index}"
  path               = "/test-path/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role" "test_nomatch" {
  name               = "${var.rName}-nomatch"
  path               = "/wrong-path/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

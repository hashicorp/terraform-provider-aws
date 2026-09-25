# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/${var.rName}"

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}

resource "aws_iam_openid_connect_provider_client_id" "test" {
  openid_connect_provider_arn = aws_iam_openid_connect_provider.test.arn
  client_id                   = "sts.amazonaws.com"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

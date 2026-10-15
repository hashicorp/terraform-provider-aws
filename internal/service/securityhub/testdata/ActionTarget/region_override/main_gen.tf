# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_account" "test" {
  region = var.region

}

resource "aws_securityhub_action_target" "test" {
  region = var.region

  depends_on  = [aws_securityhub_account.test]
  description = "description1"
  identifier  = "testaction"
  name        = "Test action"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

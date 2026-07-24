# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_finding_aggregator" "test" {
  region = var.region

  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account.test]
}

resource "aws_securityhub_account" "test" {
  region = var.region

}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

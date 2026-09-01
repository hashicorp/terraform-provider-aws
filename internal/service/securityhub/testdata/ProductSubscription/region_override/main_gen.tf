# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_securityhub_product_subscription" "test" {
  region = var.region

  depends_on  = [aws_securityhub_account.test]
  product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.region}::product/aws/guardduty"
}

data "aws_region" "current" {
  region = var.region

}
data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
  region = var.region

}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

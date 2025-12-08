# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_servicecatalog_portfolio_status" "test" {
  region = var.region

  status = "Enabled"
}


variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

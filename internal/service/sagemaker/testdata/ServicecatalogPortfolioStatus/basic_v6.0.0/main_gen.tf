# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_servicecatalog_portfolio_status" "test" {
  status = "Enabled"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.0.0"
    }
  }
}

provider "aws" {}

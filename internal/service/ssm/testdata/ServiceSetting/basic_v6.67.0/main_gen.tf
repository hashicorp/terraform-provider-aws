# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssm_service_setting" "test" {
  setting_id    = "/ssm/parameter-store/high-throughput-enabled"
  setting_value = "false"
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.67.0"
    }
  }
}

provider "aws" {}

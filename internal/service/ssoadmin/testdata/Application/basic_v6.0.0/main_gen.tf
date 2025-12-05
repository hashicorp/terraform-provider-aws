# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_ssoadmin_application" "test" {
  name                     = var.rName
  application_provider_arn = local.test_application_provider_arn
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

data "aws_ssoadmin_instances" "test" {}

locals {
  test_application_provider_arn = "arn:aws:sso::aws:applicationProvider/custom"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
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

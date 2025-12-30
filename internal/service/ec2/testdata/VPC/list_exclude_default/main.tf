# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

# tflint-ignore: terraform_unused_declarations
data "aws_vpc" "default" {
  default = true
}

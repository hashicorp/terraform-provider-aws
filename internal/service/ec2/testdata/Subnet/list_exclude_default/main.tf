# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(data.aws_vpc.default.cidr_block, 12, 4000)
  vpc_id     = data.aws_vpc.default.id
}

data "aws_vpc" "default" {
  default = true
}

# tflint-ignore: terraform_unused_declarations
data "aws_subnets" "defaults" {
  filter {
    name   = "default-for-az"
    values = ["true"]
  }
}
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_subnet" "test" {
  count = length(aws_vpc.test)

  region = var.region

  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test[count.index].id
}

resource "aws_vpc" "test" {
  count = 3

  region = var.region

  cidr_block = "10.1.0.0/16"
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

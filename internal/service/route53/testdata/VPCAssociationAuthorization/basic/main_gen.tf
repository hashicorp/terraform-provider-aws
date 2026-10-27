# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_route53_vpc_association_authorization" "test" {
  zone_id = aws_route53_zone.test.id
  vpc_id  = aws_vpc.alternate.id
}

resource "aws_vpc" "alternate" {
  provider             = "awsalternate"
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 0)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

resource "aws_route53_zone" "test" {
  name = "${var.rName}.example.com"

  vpc {
    vpc_id = aws_vpc.test.id
  }
}

resource "aws_vpc" "test" {
  cidr_block           = cidrsubnet("10.0.0.0/8", 8, 1)
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = var.rName
  }
}

provider "awsalternate" {
  access_key = var.AWS_ALTERNATE_ACCESS_KEY_ID
  profile    = var.AWS_ALTERNATE_PROFILE
  secret_key = var.AWS_ALTERNATE_SECRET_ACCESS_KEY
}

variable "AWS_ALTERNATE_ACCESS_KEY_ID" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_PROFILE" {
  type     = string
  nullable = true
  default  = null
}

variable "AWS_ALTERNATE_SECRET_ACCESS_KEY" {
  type     = string
  nullable = true
  default  = null
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

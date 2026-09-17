# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 80
}

resource "aws_lb_target_group" "test" {
  name     = var.rName
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.test[0].id
}

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 1)

resource "aws_subnet" "test" {
  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

# acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI

# acctest.configLatestAmazonLinux2HVMEBSAMI("x86_64")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-x86_64" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }
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
      version = "6.33.0"
    }
  }
}

provider "aws" {}

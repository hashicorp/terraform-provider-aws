# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ami_launch_permission" "test" {
  region = var.region

  account_id = data.aws_caller_identity.current.account_id
  image_id   = aws_ami_copy.test.id
}

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-x86_64" {
  region = var.region

  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-minimal-hvm-*-x86_64-ebs"]
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {
  region = var.region
}

resource "aws_ami_copy" "test" {
  region = var.region

  description       = var.rName
  name              = var.rName
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.region
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

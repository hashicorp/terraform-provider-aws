# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ami_launch_permission" "test" {
  count  = var.resource_count
  region = var.region

  account_id = data.aws_caller_identity.current.account_id
  image_id   = aws_ami_copy.test[count.index].id
}

data "aws_caller_identity" "current" {}

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-x86_64" {
  region = var.region

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

resource "aws_ami_copy" "test" {
  count  = var.resource_count
  region = var.region

  description       = "${var.rName}-${count.index}"
  name              = "${var.rName}-${count.index}"
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = var.region
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "The number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}

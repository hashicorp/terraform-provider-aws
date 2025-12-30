# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "aws" {}

resource "aws_autoscaling_group" "test" {
  name               = var.rName
  availability_zones = [data.aws_availability_zones.available.names[0]]

  max_size         = 1
  min_size         = 1
  desired_capacity = 1

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }

  tag {
    key                 = "test-filter"
    value               = var.rName
    propagate_at_launch = true
  }
}

resource "aws_launch_template" "test" {
  name          = var.rName
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id
  instance_type = "t4g.nano"
}

# acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI

# acctest.configLatestAmazonLinux2HVMEBSAMI("arm64")

data "aws_ami" "amzn2-ami-minimal-hvm-ebs-arm64" {
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
    values = ["arm64"]
  }
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude()

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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

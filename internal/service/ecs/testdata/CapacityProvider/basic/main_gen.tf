# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_capacity_provider" "test" {
  name = var.rName

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}

# testAccCapacityProviderConfig_base

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  name          = var.rName
}

resource "aws_autoscaling_group" "test" {
  availability_zones = data.aws_availability_zones.available.names
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = var.rName

  launch_template {
    id = aws_launch_template.test.id
  }

  tag {
    key                 = "Name"
    value               = var.rName
    propagate_at_launch = true
  }

  lifecycle {
    ignore_changes = [
      tag,
    ]
  }
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

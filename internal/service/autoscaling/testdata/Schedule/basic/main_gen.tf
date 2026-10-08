# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_autoscaling_schedule" "test" {
  scheduled_action_name  = "${var.rName}-schedule"
  min_size               = 0
  max_size               = 1
  desired_capacity       = 0
  start_time             = var.startTime
  end_time               = var.endTime
  autoscaling_group_name = aws_autoscaling_group.test.name
}

resource "aws_autoscaling_group" "test" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "${var.rName}-group"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.test.name

  tag {
    key                 = "Name"
    value               = var.rName
    propagate_at_launch = true
  }
}

resource "aws_launch_configuration" "test" {
  name_prefix   = var.rName
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  lifecycle {
    create_before_destroy = true
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

variable "startTime" {
  description = "Schedule start time"
  type        = string
  nullable    = false
}

variable "endTime" {
  description = "Schedule end time"
  type        = string
  nullable    = false
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

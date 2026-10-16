# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_ec2_instance_event_window" "test" {
  region = var.region

  name = var.rName

  time_ranges {
    start_week_day = "sunday"
    start_hour     = 2
    end_week_day   = "sunday"
    end_hour       = 6
  }
}

resource "aws_ec2_instance_event_window_association" "test" {
  region = var.region

  instance_event_window_id = aws_ec2_instance_event_window.test.id

  association_target {
    instance_ids = [aws_instance.test.id]
  }
}

# acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI

# acctest.configLatestAmazonLinux2HVMEBSAMI("x86_64")

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

data "aws_ec2_instance_type_offering" "available" {
  region = var.region

  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro", "t1.micro", "m1.small"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro", "t1.micro", "m1.small"]
}

resource "aws_instance" "test" {
  region = var.region

  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
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

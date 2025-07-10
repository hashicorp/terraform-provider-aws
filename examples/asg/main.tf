# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

locals {
  availability_zones = split(",", var.availability_zones)
}

resource "aws_elb" "web-elb" {
  name = "terraform-example-elb"

  # The same availability zone as our instances
  availability_zones = local.availability_zones

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "HTTP:80/"
    interval            = 30
  }
}

resource "aws_autoscaling_group" "web-asg" {
  availability_zones = local.availability_zones
  name               = "terraform-example-asg"
  max_size           = var.asg_max
  min_size           = var.asg_min
  desired_capacity   = var.asg_desired
  force_delete       = true
  launch_template {
    id      = aws_launch_template.web-lt.id
    version = aws_launch_template.web-lt.latest_version
  }
  load_balancers = [aws_elb.web-elb.name]

  #vpc_zone_identifier = ["${split(",", var.availability_zones)}"]
  tag {
    key                 = "Name"
    value               = "web-asg"
    propagate_at_launch = "true"
  }
}

resource "aws_launch_template" "web-lt" {
  name          = "terraform-example-lt"
  image_id      = var.aws_amis[var.aws_region]
  instance_type = var.instance_type

  # Security group
  vpc_security_group_ids = [aws_security_group.default.id]
  user_data              = base64encode(file("userdata.sh"))
  key_name               = var.key_name
}

# Our default security group to access
# the instances over SSH and HTTP
resource "aws_security_group" "default" {
  name        = "terraform_example_sg"
  description = "Used in the terraform"

  # SSH access from anywhere
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTP access from anywhere
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # outbound internet access
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

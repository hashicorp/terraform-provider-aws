# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "aws_lb_target_group" "test" {
  arn = aws_lb_target_group.test.arn
}

resource "aws_lb_target_group" "test" {
  name     = var.rName
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id

  deregistration_delay = 200
  slow_start           = 0

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 10000
  }

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = var.resource_tags
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}

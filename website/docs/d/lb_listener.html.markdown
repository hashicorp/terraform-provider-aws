---
layout: "aws"
page_title: "AWS: aws_lb_listener"
sidebar_current: "docs-aws-datasource-lb-listener"
description: |-
  Provides a Load Balancer Listener data source.
---

# Data Source: aws_lb_listener

~> **Note:** `aws_alb_listener` is known as `aws_lb_listener`. The functionality is identical.

Provides information about a Load Balancer Listener.

This data source can prove useful when a module accepts an LB Listener as an
input variable and needs to know the LB it is attached to, or other
information specific to the listener in question.

## Example Usage

```hcl
# get listener from listener arn

variable "listener_arn" {
  type = "string"
}

data "aws_lb_listener" "listener" {
  arn = "${var.listener_arn}"
}

# get listener from load_balancer_arn and port

data "aws_lb" "selected" {
  name = "default-public"
}

data "aws_lb_listener" "selected443" {
  load_balancer_arn = "${data.aws_lb.selected.arn}"
  port = 443
}

# register target group to listener

data "aws_vpc" "default" {
  default = true
}

resource "aws_lb_target_group" "core" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${data.aws_vpc.default.id}"
}

resource "aws_lb_listener_rule" "core" {
  listener_arn = "${data.aws_lb_listener.selected443.arn}"
  priority     = "${data.aws_lb_listener.selected443.next_rule_priority}"

  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.core.arn}"
  }

  condition {
    field  = "path-pattern"
    values = ["/core/*"]
  }

  # ignore priority change since it's dynamic
  lifecycle {
    ignore_changes = ["priority"]
  }
}

resource "aws_lb_target_group" "side" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${data.aws_vpc.default.id}"
}

resource "aws_lb_listener_rule" "side" {
  listener_arn = "${data.aws_lb_listener.selected443.arn}"
  priority     = "${data.aws_lb_listener.selected443.last_rule_priority + 11}"

  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.side.arn}"
  }

  condition {
    field  = "path-pattern"
    values = ["/side/*"]
  }

  # ignore priority change since it's dynamic
  lifecycle {
    ignore_changes = ["priority"]
  }
}

```

## Argument Reference

The following arguments are supported:

* `arn` - (Optional) The arn of the listener. Required if `load_balancer_arn` and `port` is not set.
* `load_balancer_arn` - (Optional) The arn of the load balander. Required if `arn` is not set.
* `port` - (Optional) The port of the listener. Required if `arn` is not set.

## Attributes Reference

See the [LB Listener Resource](/docs/providers/aws/r/lb_listener.html) for details on the returned attributes - they are identical.

Additionally, the following attributes are exported:

* `last_rule_priority` - The last used rule priority for the listener.
* `next_rule_priority` - The next available rule priority for the listener.

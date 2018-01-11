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
variable "listener_arn" {
  type = "string"
}

data "aws_lb_listener" "listener" {
  arn = "${var.listener_arn}"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) The ARN of the listener.

## Attributes Reference

See the [LB Listener Resource](/docs/providers/aws/r/lb_listener.html) for details
on the returned attributes - they are identical.

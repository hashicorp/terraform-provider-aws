---
layout: "aws"
page_title: "AWS: aws_alb_target_group"
sidebar_current: "docs-aws-datasource-alb-target-group"
description: |-
  Provides an Application Load Balancer Target Group data source.
---

# aws\_alb\_target\_group

Provides information about an Application Load Balancer Target Group.

This data source can prove useful when a module accepts an ALB Target Group as an
input variable and needs to know its attributes. It can also be used to get the ARN of
an ALB Target Group for use in other resources, given ALB Target Group name.

## Example Usage

```hcl
variable "alb_tg_arn" {
  type    = "string"
  default = ""
}

variable "alb_tg_name" {
  type    = "string"
  default = ""
}

data "aws_alb_target_group" "test" {
  arn  = "${var.alb_tg_arn}"
  name = "${var.alb_tg_name}"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Optional) The full ARN of the target group.
* `name` - (Optional) The unique name of the target group.

~> **NOTE**: When both `arn` and `name` are specified, `arn` takes precedence.

## Attributes Reference

See the [ALB Target Group Resource](/docs/providers/aws/r/alb_target_group.html) for details
on the returned attributes - they are identical.

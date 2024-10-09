---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_target_group"
description: |-
  Provides a Load Balancer Target Group data source.
---

# Data Source: aws_lb_target_group

~> **Note:** `aws_alb_target_group` is known as `aws_lb_target_group`. The functionality is identical.

Provides information about a Load Balancer Target Group.

This data source can prove useful when a module accepts an LB Target Group as an
input variable and needs to know its attributes. It can also be used to get the ARN of
an LB Target Group for use in other resources, given LB Target Group name.

## Example Usage

```terraform
variable "lb_tg_arn" {
  type    = string
  default = ""
}

variable "lb_tg_name" {
  type    = string
  default = ""
}

data "aws_lb_target_group" "test" {
  arn  = var.lb_tg_arn
  name = var.lb_tg_name
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) Full ARN of the target group.
* `name` - (Optional) Unique name of the target group.
* `tags` - (Optional) Mapping of tags, each pair of which must exactly match a pair on the desired target group.

~> **NOTE:** When both `arn` and `name` are specified, `arn` takes precedence. `tags` has the lowest precedence.

## Attribute Reference

See the [LB Target Group Resource](/docs/providers/aws/r/lb_target_group.html) for details
on the returned attributes - they are identical.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

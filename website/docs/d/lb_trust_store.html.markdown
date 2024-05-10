---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_trust_store"
description: |-
  Provides a Load Balancer Trust Store data source.
---

# Data Source: aws_lb_trust_store

~> **Note:** `aws_alb_trust_store` is known as `aws_lb_trust_store`. The functionality is identical.

Provides information about a Load Balancer Trust Store.

This data source can prove useful when a module accepts an LB Trust Store as an
input variable and needs to know its attributes. It can also be used to get the ARN of
an LB Trust Store for use in other resources, given LB Trust Store name.

## Example Usage

```terraform
variable "lb_ts_arn" {
  type    = string
  default = ""
}

variable "lb_ts_name" {
  type    = string
  default = ""
}

data "aws_lb_trust_store" "test" {
  arn  = var.lb_ts_arn
  name = var.lb_ts_name
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) Full ARN of the trust store.
* `name` - (Optional) Unique name of the trust store.

~> **NOTE:** When both `arn` and `name` are specified, `arn` takes precedence.

## Attribute Reference

See the [LB Trust Store Resource](/docs/providers/aws/r/lb_trust_store.html) for details
on the returned attributes - they are identical.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

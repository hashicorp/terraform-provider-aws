---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb"
description: |-
  Provides a Load Balancer data source.
---

# Data Source: aws_lb

~> **Note:** `aws_alb` is known as `aws_lb`. The functionality is identical.

Provides information about a Load Balancer.

This data source can prove useful when a module accepts an LB as an input
variable and needs to, for example, determine the security groups associated
with it, etc.

## Example Usage

```terraform
variable "lb_arn" {
  type    = string
  default = ""
}

variable "lb_name" {
  type    = string
  default = ""
}

data "aws_lb" "test" {
  arn  = var.lb_arn
  name = var.lb_name
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arn` - (Optional) Full ARN of the load balancer.
* `name` - (Optional) Unique name of the load balancer.
* `tags` - (Optional) Mapping of tags, each pair of which must exactly match a pair on the desired load balancer.

~> **NOTE:** When both `arn` and `name` are specified, `arn` takes precedence. `tags` has lowest precedence.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [LB Resource](/docs/providers/aws/r/lb.html) for details on the
returned attributes - they are identical.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

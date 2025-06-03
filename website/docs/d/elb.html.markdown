---
subcategory: "ELB Classic"
layout: "aws"
page_title: "AWS: aws_elb"
description: |-
  Provides a classic Elastic Load Balancer data source.
---

# Data Source: aws_elb

Provides information about a "classic" Elastic Load Balancer (ELB).
See [LB Data Source](/docs/providers/aws/d/lb.html) if you are looking for "v2"
Application Load Balancer (ALB) or Network Load Balancer (NLB).

This data source can prove useful when a module accepts an LB as an input
variable and needs to, for example, determine the security groups associated
with it, etc.

## Example Usage

```terraform
variable "lb_name" {
  type    = string
  default = ""
}

data "aws_elb" "test" {
  name = var.lb_name
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Unique name of the load balancer.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [ELB Resource](/docs/providers/aws/r/elb.html) for details on the
returned attributes - they are identical.

---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_stickiness_policy"
description: |-
  Configures Session Stickiness for a Lightsail Load Balancer
---

# Resource: aws_lightsail_lb_stickiness_policy

Configures Session Stickiness for a Lightsail Load Balancer.

## Example Usage

```terraform
resource "aws_lightsail_lb" "test" {
  name              = "test-load-balancer"
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    foo = "bar"
  }
}

resource "aws_lightsail_lb_stickiness_policy" "test" {
  lb_name         = aws_lightsail_lb.test.name
  cookie_duration = 900
  enabled         = true
}
```

## Argument Reference

This resource supports the following arguments:

* `lb_name` - (Required) The name of the load balancer to which you want to enable session stickiness.
* `cookie_duration` - (Required) The cookie duration in seconds. This determines the length of the session stickiness.
* `enabled` - (Required) - The Session Stickiness state of the load balancer. `true` to activate session stickiness or `false` to deactivate session stickiness.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name used for this load balancer (matches `lb_name`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb_stickiness_policy` using the `lb_name` attribute. For example:

```terraform
import {
  to = aws_lightsail_lb_stickiness_policy.test
  id = "example-load-balancer"
}
```

Using `terraform import`, import `aws_lightsail_lb_stickiness_policy` using the `lb_name` attribute. For example:

```console
% terraform import aws_lightsail_lb_stickiness_policy.test example-load-balancer
```

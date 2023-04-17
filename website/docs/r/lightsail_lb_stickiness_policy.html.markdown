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

The following arguments are supported:

* `lb_name` - (Required) The name of the load balancer to which you want to enable session stickiness.
* `cookie_duration` - (Required) The cookie duration in seconds. This determines the length of the session stickiness.
* `enabled` - (Required) - The Session Stickiness state of the load balancer. `true` to activate session stickiness or `false` to deactivate session stickiness.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name used for this load balancer (matches `lb_name`).

## Import

`aws_lightsail_lb_stickiness_policy` can be imported by using the `lb_name` attribute, e.g.,

```
$ terraform import aws_lightsail_lb_stickiness_policy.test example-load-balancer
```

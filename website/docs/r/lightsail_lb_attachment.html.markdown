---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_attachment"
description: |-
  Attaches a Lightsail Instance to a Lightsail Load Balancer
---

# Resource: aws_lightsail_lb_attachment

Attaches a Lightsail Instance to a Lightsail Load Balancer.

## Example Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_lb" "test" {
  name              = "test-load-balancer"
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    foo = "bar"
  }
}

resource "aws_lightsail_instance" "test" {
  name              = "test-instance"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_lb_attachment" "test" {
  lb_name       = aws_lightsail_lb.test.name
  instance_name = aws_lightsail_instance.test.name
}
```

## Argument Reference

The following arguments are supported:

* `lb_name` - (Required) The name of the Lightsail load balancer.
* `instance_name` - (Required) The name of the instance to attach to the load balancer.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of attributes to create a unique id: `lb_name`,`instance_name`

## Import

`aws_lightsail_lb_attachment` can be imported by using the name attribute, e.g.,

```
$ terraform import aws_lightsail_lb_attachment.test example-load-balancer,example-instance
```

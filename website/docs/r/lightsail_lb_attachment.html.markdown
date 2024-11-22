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
  blueprint_id      = "amazon_linux_2"
  bundle_id         = "nano_3_0"
}

resource "aws_lightsail_lb_attachment" "test" {
  lb_name       = aws_lightsail_lb.test.name
  instance_name = aws_lightsail_instance.test.name
}
```

## Argument Reference

This resource supports the following arguments:

* `lb_name` - (Required) The name of the Lightsail load balancer.
* `instance_name` - (Required) The name of the instance to attach to the load balancer.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of attributes to create a unique id: `lb_name`,`instance_name`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb_attachment` using the name attribute. For example:

```terraform
import {
  to = aws_lightsail_lb_attachment.test
  id = "example-load-balancer,example-instance"
}
```

Using `terraform import`, import `aws_lightsail_lb_attachment` using the name attribute. For example:

```console
% terraform import aws_lightsail_lb_attachment.test example-load-balancer,example-instance
```

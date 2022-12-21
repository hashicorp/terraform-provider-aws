---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_certificate_attachment"
description: |-
  Attaches a Lightsail Load Balancer Certificate to a Lightsail Load Balancer
---

# Resource: aws_lightsail_lb_certificate_attachment

Attaches a Lightsail Load Balancer Certificate to a Lightsail Load Balancer.

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

resource "aws_lightsail_lb_certificate" "test" {
  name        = "test-load-balancer-certificate"
  lb_name     = aws_lightsail_lb.test.id
  domain_name = "test.com"
}

resource "aws_lightsail_lb_certificate_attachment" "test" {
  lb_name          = aws_lightsail_lb.test.name
  certificate_name = aws_lightsail_lb_certificate.test.name
}
```

## Argument Reference

The following arguments are supported:

* `lb_name` - (Required) The name of the load balancer to which you want to associate the SSL/TLS certificate.
* `certificate_name` - (Required) The name of your SSL/TLS certificate.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A combination of attributes to create a unique id: `lb_name`,`certificate_name`

## Import

`aws_lightsail_lb_certificate_attachment` can be imported by using the name attribute, e.g.,

```
$ terraform import aws_lightsail_lb_certificate_attachment.test example-load-balancer,example-certificate
```

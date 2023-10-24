---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_https_redirection_policy"
description: |-
  Configures Https Redirection for a Lightsail Load Balancer
---

# Resource: aws_lightsail_lb_https_redirection_policy

Configures Https Redirection for a Lightsail Load Balancer. A valid Certificate must be attached to the load balancer in order to enable https redirection.

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

resource "aws_lightsail_lb_https_redirection_policy" "test" {
  lb_name = aws_lightsail_lb.test.name
  enabled = true
}
```

## Argument Reference

This resource supports the following arguments:

* `lb_name` - (Required) The name of the load balancer to which you want to enable http to https redirection.
* `enabled` - (Required) - The Https Redirection state of the load balancer. `true` to activate http to https redirection or `false` to deactivate http to https redirection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name used for this load balancer (matches `lb_name`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb_https_redirection_policy` using the `lb_name` attribute. For example:

```terraform
import {
  to = aws_lightsail_lb_https_redirection_policy.test
  id = "example-load-balancer"
}
```

Using `terraform import`, import `aws_lightsail_lb_https_redirection_policy` using the `lb_name` attribute. For example:

```console
% terraform import aws_lightsail_lb_https_redirection_policy.test example-load-balancer
```

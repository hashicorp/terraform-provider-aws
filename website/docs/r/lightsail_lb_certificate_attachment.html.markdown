---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_certificate_attachment"
description: |-
  Manages a Lightsail Load Balancer Certificate attachment to a Lightsail Load Balancer.
---

# Resource: aws_lightsail_lb_certificate_attachment

Manages a Lightsail Load Balancer Certificate attachment to a Lightsail Load Balancer.

Use this resource to attach a validated SSL/TLS certificate to a Lightsail Load Balancer to enable HTTPS traffic. The certificate must be validated before it can be attached to the load balancer.

## Example Usage

```terraform
resource "aws_lightsail_lb" "example" {
  name              = "example-load-balancer"
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    foo = "bar"
  }
}

resource "aws_lightsail_lb_certificate" "example" {
  name        = "example-load-balancer-certificate"
  lb_name     = aws_lightsail_lb.example.id
  domain_name = "example.com"
}

resource "aws_lightsail_lb_certificate_attachment" "example" {
  lb_name          = aws_lightsail_lb.example.name
  certificate_name = aws_lightsail_lb_certificate.example.name
}
```

## Argument Reference

The following arguments are required:

* `certificate_name` - (Required) Name of your SSL/TLS certificate.
* `lb_name` - (Required) Name of the load balancer to which you want to associate the SSL/TLS certificate.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Combination of attributes to create a unique id: `lb_name`,`certificate_name`

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb_certificate_attachment` using the name attribute. For example:

```terraform
import {
  to = aws_lightsail_lb_certificate_attachment.example
  id = "example-load-balancer,example-certificate"
}
```

Using `terraform import`, import `aws_lightsail_lb_certificate_attachment` using the name attribute. For example:

```console
% terraform import aws_lightsail_lb_certificate_attachment.example example-load-balancer,example-certificate
```

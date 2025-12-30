---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_certificate"
description: |-
  Manages a Lightsail Load Balancer Certificate.
---

# Resource: aws_lightsail_lb_certificate

Manages a Lightsail Load Balancer Certificate.

Use this resource to create and manage SSL/TLS certificates for Lightsail Load Balancers. The certificate must be validated before it can be attached to a load balancer to enable HTTPS traffic.

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
```

## Argument Reference

The following arguments are required:

* `domain_name` - (Required) Domain name (e.g., example.com) for your SSL/TLS certificate.
* `lb_name` - (Required) Load balancer name where you want to create the SSL/TLS certificate.
* `name` - (Required) SSL/TLS certificate name.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `subject_alternative_names` - (Optional) Set of domains that should be SANs in the issued certificate. `domain_name` attribute is automatically added as a Subject Alternative Name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the lightsail certificate.
* `created_at` - Timestamp when the instance was created.
* `domain_validation_records` - Set of domain validation objects which can be used to complete certificate validation. Can have more than one element, e.g., if SANs are defined.
* `id` - Combination of attributes to create a unique id: `lb_name`,`name`
* `support_code` - Support code for the certificate.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb_certificate` using the id attribute. For example:

```terraform
import {
  to = aws_lightsail_lb_certificate.example
  id = "example-load-balancer,example-load-balancer-certificate"
}
```

Using `terraform import`, import `aws_lightsail_lb_certificate` using the id attribute. For example:

```console
% terraform import aws_lightsail_lb_certificate.example example-load-balancer,example-load-balancer-certificate
```

---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb_stickiness_policy"
description: |-
  Manages session stickiness for a Lightsail Load Balancer.
---

# Resource: aws_lightsail_lb_stickiness_policy

Manages session stickiness for a Lightsail Load Balancer.

Use this resource to configure session stickiness to ensure that user sessions are consistently routed to the same backend instance. This helps maintain session state for applications that store session data locally on the server.

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

resource "aws_lightsail_lb_stickiness_policy" "example" {
  lb_name         = aws_lightsail_lb.example.name
  cookie_duration = 900
  enabled         = true
}
```

## Argument Reference

The following arguments are required:

* `cookie_duration` - (Required) Cookie duration in seconds. This determines the length of the session stickiness.
* `enabled` - (Required) Whether to enable session stickiness for the load balancer.
* `lb_name` - (Required) Name of the load balancer to which you want to enable session stickiness.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name used for this load balancer (matches `lb_name`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb_stickiness_policy` using the `lb_name` attribute. For example:

```terraform
import {
  to = aws_lightsail_lb_stickiness_policy.example
  id = "example-load-balancer"
}
```

Using `terraform import`, import `aws_lightsail_lb_stickiness_policy` using the `lb_name` attribute. For example:

```console
% terraform import aws_lightsail_lb_stickiness_policy.example example-load-balancer
```

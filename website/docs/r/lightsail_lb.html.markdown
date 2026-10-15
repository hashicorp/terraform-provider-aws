---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb"
description: |-
  Manages a Lightsail Load Balancer.
---

# Resource: aws_lightsail_lb

Manages a Lightsail load balancer resource.

Use this resource to distribute incoming traffic across multiple Lightsail instances to improve application availability and performance.

## Example Usage

```terraform
resource "aws_lightsail_lb" "example" {
  name              = "example-load-balancer"
  health_check_path = "/"
  instance_port     = 80
  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_port` - (Required) Instance port the load balancer will connect to.
* `name` - (Required) Name of the Lightsail load balancer.

The following arguments are optional:

* `health_check_path` - (Optional) Health check path of the load balancer. Default value `/`.
* `ip_address_type` - (Optional) IP address type of the load balancer. Valid values: `dualstack`, `ipv4`. Default value `dualstack`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Lightsail load balancer.
* `created_at` - Timestamp when the load balancer was created.
* `dns_name` - DNS name of the load balancer.
* `id` - Name used for this load balancer (matches `name`).
* `protocol` - Protocol of the load balancer.
* `public_ports` - Public ports of the load balancer.
* `support_code` - Support code for the load balancer. Include this code in your email to support when you have questions about a load balancer in Lightsail. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_lb` using the name attribute. For example:

```terraform
import {
  to = aws_lightsail_lb.example
  id = "example-load-balancer"
}
```

Using `terraform import`, import `aws_lightsail_lb` using the name attribute. For example:

```console
% terraform import aws_lightsail_lb.example example-load-balancer
```

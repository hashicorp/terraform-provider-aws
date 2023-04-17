---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_lb"
description: |-
  Provides a Lightsail Load Balancer
---

# Resource: aws_lightsail_lb

Creates a Lightsail load balancer resource.

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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail load balancer.
* `instance_port` - (Required) The instance port the load balancer will connect.
* `health_check_path` - (Optional) The health check path of the load balancer. Default value "/".
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name used for this load balancer (matches `name`).
* `arn` - The ARN of the Lightsail load balancer.
* `created_at` - The timestamp when the load balancer was created.
* `dns_name` - The DNS name of the load balancer.
* `protocol` - The protocol of the load balancer.
* `public_ports` - The public ports of the load balancer.
* `support_code` - The support code for the database. Include this code in your email to support when you have questions about a database in Lightsail. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`aws_lightsail_lb` can be imported by using the name attribute, e.g.,

```
$ terraform import aws_lightsail_lb.test example-load-balancer
```

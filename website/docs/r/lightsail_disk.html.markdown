---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_disk"
description: |-
  Provides a Lightsail Disk resource
---

# Resource: aws_lightsail_disk

Provides a Lightsail Disk resource.

## Example Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_disk" "test" {
  name              = "test"
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the Lightsail load balancer.
* `size_in_gb` - (Required) The instance port the load balancer will connect.
* `availability_zone` - (Required) The Availability Zone in which to create your disk.
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the disk  (matches `name`).
* `arn` - The ARN of the Lightsail load balancer.
* `created_at` - The timestamp when the load balancer was created.
* `support_code` - The support code for the disk. Include this code in your email to support when you have questions about a disk in Lightsail. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_disk` using the name attribute. For example:

```terraform
import {
  to = aws_lightsail_disk.test
  id = "test"
}
```

Using `terraform import`, import `aws_lightsail_disk` using the name attribute. For example:

```console
% terraform import aws_lightsail_disk.test test
```

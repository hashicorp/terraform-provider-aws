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

The following arguments are supported:

* `name` - (Required) The name of the Lightsail load balancer.
* `size_in_gb` - (Required) The instance port the load balancer will connect.
* `availability_zone` - (Required) The Availability Zone in which to create your disk.
* `tags` - (Optional) A map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the disk  (matches `name`).
* `arn` - The ARN of the Lightsail load balancer.
* `created_at` - The timestamp when the load balancer was created.
* `support_code` - The support code for the disk. Include this code in your email to support when you have questions about a disk in Lightsail. This code enables our support team to look up your Lightsail information more easily.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`aws_lightsail_disk` can be imported by using the name attribute, e.g.,

```shell
$ terraform import aws_lightsail_disk.test test
```

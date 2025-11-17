---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_disk"
description: |-
  Manages a Lightsail block storage disk.
---

# Resource: aws_lightsail_disk

Manages a Lightsail disk. Use this resource to create additional block storage that can be attached to Lightsail instances for extra storage capacity.

## Example Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_lightsail_disk" "example" {
  name              = "example-disk"
  size_in_gb        = 8
  availability_zone = data.aws_availability_zones.available.names[0]
}
```

## Argument Reference

The following arguments are required:

* `availability_zone` - (Required) Availability Zone in which to create the disk.
* `name` - (Required) Name of the disk. Must begin with an alphabetic character and contain only alphanumeric characters, underscores, hyphens, and dots.
* `size_in_gb` - (Required) Size of the disk in GB.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. To create a key-only tag, use an empty string as the value. If configured with a provider `default_tags` configuration block present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the disk.
* `created_at` - Date and time when the disk was created.
* `id` - Name of the disk (matches `name`).
* `support_code` - Support code for the disk. Include this code in your email to support when you have questions about a disk in Lightsail.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider `default_tags` configuration block.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_lightsail_disk` using the name attribute. For example:

```terraform
import {
  to = aws_lightsail_disk.example
  id = "example-disk"
}
```

Using `terraform import`, import `aws_lightsail_disk` using the name attribute. For example:

```console
% terraform import aws_lightsail_disk.example example-disk
```

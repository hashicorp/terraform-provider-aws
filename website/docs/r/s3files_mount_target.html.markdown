---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_mount_target"
description: |-
  Terraform resource for managing an Amazon S3 Files Mount Target.
---

# Resource: aws_s3files_mount_target

Terraform resource for managing an Amazon S3 Files Mount Target.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3files_mount_target" "example" {
  file_system_id = aws_s3files_file_system.example.file_system_id
  subnet_id      = aws_subnet.example.id
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required, Forces new resource) The ID of the S3 file system.
* `subnet_id` - (Required, Forces new resource) The ID of the subnet where the mount target will be created.

The following arguments are optional:

* `ipv4_address` - (Optional, Forces new resource) A specific IPv4 address to assign to the mount target.
* `ipv6_address` - (Optional, Forces new resource) A specific IPv6 address to assign to the mount target.
* `security_groups` - (Optional) An array of VPC security group IDs to associate with the mount target.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `availability_zone_id` - The Availability Zone ID where the mount target is located.
* `mount_target_id` - The ID of the mount target.
* `network_interface_id` - The ID of the network interface associated with the mount target.
* `owner_id` - The AWS account ID of the mount target owner.
* `status` - The lifecycle state of the mount target.
* `status_message` - Additional information about the mount target status.
* `vpc_id` - The ID of the VPC where the mount target is located.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files Mount Target using the `mount_target_id`. For example:

```terraform
import {
  to = aws_s3files_mount_target.example
  id = "fsmt-0123456789abcdef0"
}
```

Using `terraform import`, import S3 Files Mount Target using the `mount_target_id`. For example:

```console
% terraform import aws_s3files_mount_target.example fsmt-0123456789abcdef0
```

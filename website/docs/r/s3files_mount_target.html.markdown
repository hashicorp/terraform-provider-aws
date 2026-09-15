---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_mount_target"
description: |-
  Manages an S3 Files Mount Target.
---

# Resource: aws_s3files_mount_target

Manages an S3 Files Mount Target.

## Example Usage

```terraform
resource "aws_s3files_mount_target" "example" {
  file_system_id = aws_s3files_file_system.example.id
  subnet_id      = aws_subnet.example.id
}
```

## Argument Reference

The following arguments are required:

* `file_system_id` - (Required) File system ID. Changing this value forces replacement.
* `subnet_id` - (Required) Subnet ID. Changing this value forces replacement.

The following arguments are optional:

* `ip_address_type` - (Optional) IP address type.
* `ipv4_address` - (Optional) IPv4 address.
* `ipv6_address` - (Optional) IPv6 address.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_groups` - (Optional) Security group IDs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `availability_zone_id` - Availability Zone ID.
* `id` - Identifier of the mount target.
* `network_interface_id` - Network interface ID.
* `owner_id` - AWS account ID of the owner.
* `status` - Mount target status.
* `status_message` - Status message.
* `vpc_id` - VPC ID.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_s3files_mount_target.example
  identity = {
    id = "fsmt-1234567890abcdef0"
  }
}

resource "aws_s3files_mount_target" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - Identifier of the mount target.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import S3 Files Mount Target using the resource ID. For example:

```terraform
import {
  to = aws_s3files_mount_target.example
  id = "fsmt-1234567890abcdef0"
}
```

Using `terraform import`, import S3 Files Mount Target using `id`. For example:

```console
% terraform import aws_s3files_mount_target.example fsmt-1234567890abcdef0
```

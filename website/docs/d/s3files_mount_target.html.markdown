---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_mount_target"
description: |-
  Provides details about an S3 Files Mount Target.
---

# Data Source: aws_s3files_mount_target

Provides details about an S3 Files Mount Target.

## Example Usage

```terraform
data "aws_s3files_mount_target" "example" {
  id = "fsmt-1234567890abcdef0"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) Mount target ID.

The following arguments are optional:

* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_zone_id` - Availability Zone ID.
* `file_system_id` - File system ID.
* `ipv4_address` - IPv4 address.
* `ipv6_address` - IPv6 address.
* `network_interface_id` - Network interface ID.
* `owner_id` - AWS account ID of the owner.
* `security_groups` - Security group IDs.
* `status` - Mount target status.
* `status_message` - Status message.
* `subnet_id` - Subnet ID.
* `vpc_id` - VPC ID.

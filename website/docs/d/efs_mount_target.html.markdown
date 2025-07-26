---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_mount_target"
description: |-
  Provides an Elastic File System Mount Target (EFS) data source.
---

# Data Source: aws_efs_mount_target

Provides information about an Elastic File System Mount Target (EFS).

## Example Usage

```terraform
variable "mount_target_id" {
  type    = string
  default = ""
}

data "aws_efs_mount_target" "by_id" {
  mount_target_id = var.mount_target_id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `access_point_id` - (Optional) ID or ARN of the access point whose mount target that you want to find. It must be included if a `file_system_id` and `mount_target_id` are not included.
* `file_system_id` - (Optional) ID or ARN of the file system whose mount target that you want to find. It must be included if an `access_point_id` and `mount_target_id` are not included.
* `mount_target_id` - (Optional) ID or ARN of the mount target that you want to find. It must be included in your request if an `access_point_id` and `file_system_id` are not included.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `file_system_arn` - Amazon Resource Name of the file system for which the mount target is intended.
* `subnet_id` - ID of the mount target's subnet.
* `ip_address` - Address at which the file system may be mounted via the mount target.
* `security_groups` - List of VPC security group IDs attached to the mount target.
* `dns_name` - DNS name for the EFS file system.
* `mount_target_dns_name` - The DNS name for the given subnet/AZ per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).
* `network_interface_id` - The ID of the network interface that Amazon EFS created when it created the mount target.
* `availability_zone_name` - The name of the Availability Zone (AZ) that the mount target resides in.
* `availability_zone_id` - The unique and consistent identifier of the Availability Zone (AZ) that the mount target resides in.
* `owner_id` - AWS account ID that owns the resource.

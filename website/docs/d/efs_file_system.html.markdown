---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_file_system"
description: |-
  Provides an Elastic File System (EFS) File System data source.
---

# Data Source: aws_efs_file_system

Provides information about an Elastic File System (EFS) File System.

## Example Usage

```terraform
variable "file_system_id" {
  type    = string
  default = ""
}

data "aws_efs_file_system" "by_id" {
  file_system_id = var.file_system_id
}

data "aws_efs_file_system" "by_tag" {
  tags = {
    Environment = "dev"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `file_system_id` - (Optional) ID that identifies the file system (e.g., fs-ccfc0d65).
* `creation_token` - (Optional) Restricts the list to the file system with this creation token.
* `tags` - (Optional) Restricts the list to the file system with these tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the file system.
* `availability_zone_name` - The Availability Zone name in which the file system's One Zone storage classes exist.
* `availability_zone_id` - The identifier of the Availability Zone in which the file system's One Zone storage classes exist.
* `dns_name` - DNS name for the filesystem per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).
* `encrypted` - Whether EFS is encrypted.
* `kms_key_id` - ARN for the KMS encryption key.
* `lifecycle_policy` - File system [lifecycle policy](https://docs.aws.amazon.com/efs/latest/ug/API_LifecyclePolicy.html) object.
* `performance_mode` - File system performance mode.
* `provisioned_throughput_in_mibps` - The throughput, measured in MiB/s, that you want to provision for the file system.
* `tags` -A map of tags to assign to the file system.
* `throughput_mode` - Throughput mode for the file system.
* `size_in_bytes` - Current byte count used by the file system.

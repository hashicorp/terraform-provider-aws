---
subcategory: "EFS"
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
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Optional) The ID that identifies the file system (e.g. fs-ccfc0d65).
* `creation_token` - (Optional) Restricts the list to the file system with this creation token.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `availability_zone_name` - The Availability Zone name in which the file system's One Zone storage classes exist.
* `availability_zone_id` - The identifier of the Availability Zone in which the file system's One Zone storage classes exist.
* `dns_name` - The DNS name for the filesystem per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).
* `encrypted` - Whether EFS is encrypted.
* `kms_key_id` - The ARN for the KMS encryption key.
* `lifecycle_policy` - A file system [lifecycle policy](https://docs.aws.amazon.com/efs/latest/ug/API_LifecyclePolicy.html) object.
* `performance_mode` - The file system performance mode.
* `provisioned_throughput_in_mibps` - The throughput, measured in MiB/s, that you want to provision for the file system.
* `tags` -A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).
* `throughput_mode` - Throughput mode for the file system.
* `size_in_bytes` - The current byte count used by the file system.

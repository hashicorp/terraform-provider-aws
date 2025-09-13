---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_fsx_openzfs_file_system"
description: |-
  Manages an FSx OpenZfs Location within AWS DataSync.
---

# Resource: aws_datasync_location_fsx_openzfs_file_system

Manages an AWS DataSync FSx OpenZfs Location.

## Example Usage

```terraform
resource "aws_datasync_location_fsx_openzfs_file_system" "example" {
  fsx_filesystem_arn  = aws_fsx_openzfs_file_system.example.arn
  security_group_arns = [aws_security_group.example.arn]

  protocol {
    nfs {
      mount_options {
        version = "AUTOMATIC"
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `fsx_filesystem_arn` - (Required) The Amazon Resource Name (ARN) for the FSx for OpenZfs file system.
* `protocol` - (Required) The type of protocol that DataSync uses to access your file system. See below.
* `security_group_arns` - (Optional) The Amazon Resource Names (ARNs) of the security groups that are to use to configure the FSx for openzfs file system.
* `subdirectory` - (Optional) Subdirectory to perform actions as source or destination. Must start with `/fsx`.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### protocol

* `nfs` - (Required) Represents the Network File System (NFS) protocol that DataSync uses to access your FSx for OpenZFS file system. See below.

### nfs

* `mount_options` - (Required) Represents the mount options that are available for DataSync to access an NFS location. See below.

### mount_options

* `version` - (Optional) The specific NFS version that you want DataSync to use for mounting your NFS share. Valid values: `AUTOMATIC`, `NFS3`, `NFS4_0` and `NFS4_1`. Default: `AUTOMATIC`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `uri` - The URL of the FSx for openzfs location that was described.
* `creation_time` - The time that the FSx for openzfs location was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datasync_location_fsx_openzfs_file_system` using the `DataSync-ARN#FSx-openzfs-ARN`. For example:

```terraform
import {
  to = aws_datasync_location_fsx_openzfs_file_system.example
  id = "arn:aws:datasync:us-west-2:123456789012:location/loc-12345678901234567#arn:aws:fsx:us-west-2:123456789012:file-system/fs-08e04cd442c1bb94a"
}
```

Using `terraform import`, import `aws_datasync_location_fsx_openzfs_file_system` using the `DataSync-ARN#FSx-openzfs-ARN`. For example:

```console
% terraform import aws_datasync_location_fsx_openzfs_file_system.example arn:aws:datasync:us-west-2:123456789012:location/loc-12345678901234567#arn:aws:fsx:us-west-2:123456789012:file-system/fs-08e04cd442c1bb94a
```

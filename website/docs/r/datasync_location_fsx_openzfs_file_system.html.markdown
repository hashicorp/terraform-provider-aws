---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_fsx_openzfs_file_system"
description: |-
  Manages an FSx openzfs Location within AWS DataSync.
---

# Resource: aws_datasync_location_fsx_openzfs_file_system

Manages an AWS DataSync FSx openzfs Location.

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

The following arguments are supported:

* `fsx_filesystem_arn` - (Required) The Amazon Resource Name (ARN) for the FSx for openzfs file system.
* `protocol` - (Required) The type of protocol that DataSync uses to access your file system. See [Protocol](#protocol) Below.
* `security_group_arns` - (Optional) The Amazon Resource Names (ARNs) of the security groups that are to use to configure the FSx for openzfs file system.
* `subdirectory` - (Optional) Subdirectory to perform actions as source or destination. Must start with `/fsx`
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Protocol

* `nfs` - (Required) The type of protocol that DataSync uses to access your file system. See [nfs](#nfs).

#### NFS

* `mount_options` - (Required) The type of protocol that DataSync uses to access your file system. See [Mount Options](#mount-options).

##### Mount Options

* `version` - (Optional) The specific NFS version that you want DataSync to use for mounting your NFS share. Valid values: `AUTOMATIC`, `NFS3`, `NFS4_0` and `NFS4_1`. Default: `AUTOMATIC`


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `uri` - The URL of the FSx for openzfs location that was described.
* `creation_time` - The time that the FSx for openzfs location was created.

## Import

`aws_datasync_location_fsx_openzfs_file_system` can be imported by using the `DataSync-ARN#FSx-openzfs-ARN`, e.g.,

```
$ terraform import aws_datasync_location_fsx_openzfs_file_system.example arn:aws:datasync:us-west-2:123456789012:location/loc-12345678901234567#arn:aws:fsx:us-west-2:476956259333:file-system/fs-08e04cd442c1bb94a
```

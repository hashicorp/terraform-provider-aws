---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_openzfs_snapshot"
description: |-
  Manages an Amazon FSx for OpenZFS snapshot.
---

# Resource: aws_fsx_openzfs_snapshot

Manages an Amazon FSx for OpenZFS volume.
See the [FSx OpenZFS User Guide](https://docs.aws.amazon.com/fsx/latest/OpenZFSGuide/what-is-fsx.html) for more information.

## Example Usage

### Root volume Example

```terraform
resource "aws_fsx_openzfs_snapshot" "example" {
  name      = "example"
  volume_id = aws_fsx_openzfs_file_system.example.root_volume_id
}

resource "aws_fsx_openzfs_file_system" "example" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.example.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
}
```

### Child volume Example

```terraform
resource "aws_fsx_openzfs_snapshot" "example" {
  name      = "example"
  volume_id = aws_fsx_openzfs_volume.example.id
}

resource "aws_fsx_openzfs_volume" "example" {
  name             = "example"
  parent_volume_id = aws_fsx_openzfs_file_system.example.root_volume_id
}

resource "aws_fsx_openzfs_file_system" "example" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.example.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Snapshot. You can use a maximum of 203 alphanumeric characters plus either _ or -  or : or . for the name.
* `tags` - (Optional) A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. If you have set `copy_tags_to_backups` to true, and you specify one or more tags, no existing file system tags are copied from the file system to the backup.
* `volume_id` - (Optional) The ID of the volume to snapshot. This can be the root volume or a child volume.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the snapshot.
* `id` - Identifier of the snapshot, e.g., `fsvolsnap-12345678`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)
* `update` - (Default `30m`)

## Import

FSx OpenZFS snapshot can be imported using the `id`, e.g.,

```
$ terraform import aws_fsx_openzfs_snapshot.example fs-543ab12b1ca672f33
```

---
subcategory: "File System (FSx)"
layout: "aws"
page_title: "AWS: aws_fsx_backup"
description: |-
  Manages a FSx Backup.
---

# Resource: aws_fsx_backup

Provides a FSx Backup resource.

## Example Usage

```terraform
resource "aws_fsx_backup" "example" {
  file_system_id = aws_fsx_lustre_file_system.example.id
}

resource "aws_fsx_lustre_file_system" "example" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.example.id]
  deployment_type             = "PERSISTENT_1"
  per_unit_storage_throughput = 50
}
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Required) The ID of the file system to back up.
* `tags` - (Optional) A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. If you have set `copy_tags_to_backups` to true, and you specify one or more tags, no existing file system tags are copied from the file system to the backup.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the backup.
* `id` - Identifier of the backup, e.g., `fs-12345678`
* `kms_key_id` -  The ID of the AWS Key Management Service (AWS KMS) key used to encrypt the backup of the Amazon FSx file system's data at rest.
* `owner_id` - AWS account identifier that created the file system.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `type` - The type of the file system backup.

## Timeouts

`aws_fsx_backup` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

* `create` - (Default `10m`) How long to wait for the backup to be created.
* `delete` - (Default `10m`) How long to wait for the backup to be deleted.

## Import

FSx Backups can be imported using the `id`, e.g.,

```
$ terraform import aws_fsx_backup.example fs-543ab12b1ca672f33
```

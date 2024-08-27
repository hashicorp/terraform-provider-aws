---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_backup"
description: |-
  Manages a FSx Backup.
---

# Resource: aws_fsx_backup

Provides a FSx Backup resource.

## Example Usage

## Lustre Example

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

## Windows Example

```terraform
resource "aws_fsx_backup" "example" {
  file_system_id = aws_fsx_windows_file_system.example.id
}

resource "aws_fsx_windows_file_system" "example" {
  active_directory_id = aws_directory_service_directory.eample.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.example1.id]
  throughput_capacity = 8
}
```

## ONTAP Example

```terraform
resource "aws_fsx_backup" "example" {
  volume_id = aws_fsx_ontap_volume.example.id
}

resource "aws_fsx_ontap_volume" "example" {
  name                       = "example"
  junction_path              = "/example"
  size_in_megabytes          = 1024
  storage_efficiency_enabled = true
  storage_virtual_machine_id = aws_fsx_ontap_storage_virtual_machine.test.id
}
```

## OpenZFS Example

```terraform
resource "aws_fsx_backup" "example" {
  file_system_id = aws_fsx_openzfs_file_system.example.id
}

resource "aws_fsx_openzfs_file_system" "example" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.example.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64
}
```

## Argument Reference

This resource supports the following arguments:

Note - Only file_system_id or volume_id can be specified. file_system_id is used for Lustre and Windows, volume_id is used for ONTAP.

* `file_system_id` - (Optional) The ID of the file system to back up. Required if backing up Lustre or Windows file systems.
* `tags` - (Optional) A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level. If you have set `copy_tags_to_backups` to true, and you specify one or more tags, no existing file system tags are copied from the file system to the backup.
* `volume_id` - (Optional) The ID of the volume to back up. Required if backing up a ONTAP Volume.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the backup.
* `id` - Identifier of the backup, e.g., `fs-12345678`
* `kms_key_id` -  The ID of the AWS Key Management Service (AWS KMS) key used to encrypt the backup of the Amazon FSx file system's data at rest.
* `owner_id` - AWS account identifier that created the file system.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `type` - The type of the file system backup.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FSx Backups using the `id`. For example:

```terraform
import {
  to = aws_fsx_backup.example
  id = "fs-543ab12b1ca672f33"
}
```

Using `terraform import`, import FSx Backups using the `id`. For example:

```console
% terraform import aws_fsx_backup.example fs-543ab12b1ca672f33
```

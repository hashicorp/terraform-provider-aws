---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_replication_configuration"
description: Provides an Elastic File System (EFS) Replication Configuration.
---

# Resource: aws_efs_replication_configuration

Creates a replica of an existing EFS file system in the same or another region. Creating this resource causes the source EFS file system to be replicated to a new read-only destination EFS file system (unless using the `destination.file_system_id` attribute). Deleting this resource will cause the replication from source to destination to stop and the destination file system will no longer be read only.

~> **NOTE:** Deleting this resource does **not** delete the destination file system that was created.

## Example Usage

Will create a replica using regional storage in us-west-2 that will be encrypted by the default EFS KMS key `/aws/elasticfilesystem`.

```terraform
resource "aws_efs_file_system" "example" {}

resource "aws_efs_replication_configuration" "example" {
  source_file_system_id = aws_efs_file_system.example.id

  destination {
    region = "us-west-2"
  }
}
```

Replica will be created as One Zone storage in the us-west-2b Availability Zone and encrypted with the specified KMS key.

```terraform
resource "aws_efs_file_system" "example" {}

resource "aws_efs_replication_configuration" "example" {
  source_file_system_id = aws_efs_file_system.example.id

  destination {
    availability_zone_name = "us-west-2b"
    kms_key_id             = "1234abcd-12ab-34cd-56ef-1234567890ab"
  }
}
```

Will create a replica and set the existing file system with id `fs-1234567890` in us-west-2 as destination.

```terraform
resource "aws_efs_file_system" "example" {}

resource "aws_efs_replication_configuration" "example" {
  source_file_system_id = aws_efs_file_system.example.id

  destination {
    file_system_id = "fs-1234567890"
    region         = "us-west-2"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `destination` - (Required) A destination configuration block (documented below).
* `source_file_system_id` - (Required) The ID of the file system that is to be replicated.

### Destination Arguments

`destination` supports the following arguments:

* `availability_zone_name` - (Optional) The availability zone in which the replica should be created. If specified, the replica will be created with One Zone storage. If omitted, regional storage will be used.
* `file_system_id` - (Optional) The ID of the destination file system for the replication. If no ID is provided, then EFS creates a new file system with the default settings.
* `kms_key_id` - (Optional) The Key ID, ARN, alias, or alias ARN of the KMS key that should be used to encrypt the replica file system. If omitted, the default KMS key for EFS `/aws/elasticfilesystem` will be used.
* `region` - (Optional) The region in which the replica should be created.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_time` - When the replication configuration was created.
* `destination[0].file_system_id` - The fs ID of the replica.
* `destination[0].status` - The status of the replication.
* `original_source_file_system_arn` - The Amazon Resource Name (ARN) of the original source Amazon EFS file system in the replication configuration.
* `source_file_system_arn` - The Amazon Resource Name (ARN) of the current source file system in the replication configuration.
* `source_file_system_region` - The AWS Region in which the source Amazon EFS file system is located.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EFS Replication Configurations using the file system ID of either the source or destination file system. When importing, the `availability_zone_name` and `kms_key_id` attributes must **not** be set in the configuration. The AWS API does not return these values when querying the replication configuration and their presence will therefore show as a diff in a subsequent plan. For example:

```terraform
import {
  to = aws_efs_replication_configuration.example
  id = "fs-id"
}
```

Using `terraform import`, import EFS Replication Configurations using the file system ID of either the source or destination file system. When importing, the `availability_zone_name` and `kms_key_id` attributes must **not** be set in the configuration. The AWS API does not return these values when querying the replication configuration and their presence will therefore show as a diff in a subsequent plan. For example:

```console
% terraform import aws_efs_replication_configuration.example fs-id
```

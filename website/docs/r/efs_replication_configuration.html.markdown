---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_replication_configuration"
description: Provides an Elastic File System (EFS) Replication Configuration.
---

# Resource: aws_efs_replication_configuration

Creates a replica of an existing EFS file system in the same or another region. Creating this resource causes the source EFS file system to be replicated to a new read-only destination EFS file system. Deleting this resource will cause the replication from source to destination to stop and the destination file system will no longer be read only.

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

## Argument Reference

The following arguments are supported:

* `source_file_system_id` - (Required) The ID of the file system that is to be replicated.
* `destination` - (Required) A destination configuration block (documented below).

### Destination Arguments

For **destination** the following attributes are supported:

* `availability_zone_name` - (Optional) The availability zone in which the replica should be created. If specified, the replica will be created with One Zone storage. If omitted, regional storage will be used.
* `kms_key_id` - (Optional) The Key ID, ARN, alias, or alias ARN of the KMS key that should be used to encrypt the replica file system. If omitted, the default KMS key for EFS `/aws/elasticfilesystem` will be used.
* `region` - (Optional) The region in which the replica should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `creation_time` - When the replication configuration was created.
* `original_source_file_system_arn` - The Amazon Resource Name (ARN) of the original source Amazon EFS file system in the replication configuration.
* `source_file_system_arn` - The Amazon Resource Name (ARN) of the current source file system in the replication configuration.
* `source_file_system_region` - The AWS Region in which the source Amazon EFS file system is located.
* `destination[0].file_system_id` - The fs ID of the replica.
* `destination[0].status` - The status of the replication.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

* `create` - (Default `10 minutes`) Used when creating the replication configuration.
* `delete` - (Default `20 minutes`) Used when deleting the replication configuration.

## Import

EFS Replication Configurations can be imported using the file system ID of either the source or destination file system. When importing, the `availability_zone_name` and `kms_key_id` attributes must **not** be set in the configuration. The AWS API does not return these values when querying the replication configuration and their presence will therefore show as a diff in a subsequent plan.

```
$ terraform import aws_efs_replication_configuration.example fs-id
```

---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_snapshot_copy"
description: |-
  Manages an RDS database instance snapshot copy.
---

# Resource: aws_db_snapshot_copy

Manages an RDS database instance snapshot copy. For managing RDS database cluster snapshots, see the [`aws_db_cluster_snapshot` resource](/docs/providers/aws/r/db_cluster_snapshot.html).

## Example Usage

```terraform
resource "aws_db_instance" "example" {
  allocated_storage = 10
  engine            = "mysql"
  engine_version    = "5.6.21"
  instance_class    = "db.t2.micro"
  db_name           = "baz"
  password          = "barbarbarbar"
  username          = "foo"

  maintenance_window      = "Fri:09:00-Fri:09:30"
  backup_retention_period = 0
  parameter_group_name    = "default.mysql5.6"
}

resource "aws_db_snapshot" "example" {
  db_instance_identifier = aws_db_instance.example.identifier
  db_snapshot_identifier = "testsnapshot1234"
}

resource "aws_db_snapshot_copy" "example" {
  source_db_snapshot_identifier = aws_db_snapshot.example.db_snapshot_arn
  target_db_snapshot_identifier = "testsnapshot1234-copy"
}
```

## Argument Reference

This resource supports the following arguments:

* `copy_tags` - (Optional) Whether to copy existing tags. Defaults to `false`.
* `destination_region` - (Optional) The Destination region to place snapshot copy.
* `kms_key_id` - (Optional) KMS key ID.
* `option_group_name`- (Optional) The name of an option group to associate with the copy of the snapshot.
* `presigned_url` - (Optional) he URL that contains a Signature Version 4 signed request.
* `source_db_snapshot_identifier` - (Required) Snapshot identifier of the source snapshot.
* `target_custom_availability_zone` - (Optional) The external custom Availability Zone.
* `target_db_snapshot_identifier` - (Required) The Identifier for the snapshot.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Snapshot Identifier.
* `allocated_storage` - Specifies the allocated storage size in gigabytes (GB).
* `availability_zone` - Specifies the name of the Availability Zone the DB instance was located in at the time of the DB snapshot.
* `db_snapshot_arn` - The Amazon Resource Name (ARN) for the DB snapshot.
* `encrypted` - Specifies whether the DB snapshot is encrypted.
* `engine` - Specifies the name of the database engine.
* `engine_version` - Specifies the version of the database engine.
* `iops` - Specifies the Provisioned IOPS (I/O operations per second) value of the DB instance at the time of the snapshot.
* `kms_key_id` - The ARN for the KMS encryption key.
* `license_model` - License model information for the restored DB instance.
* `option_group_name` - Provides the option group name for the DB snapshot.
* `shared_accounts` - (Optional) List of AWS Account ids to share snapshot with, use `all` to make snaphot public.
* `source_db_snapshot_identifier` - The DB snapshot Arn that the DB snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `source_region` - The region that the DB snapshot was created in or copied from.
* `storage_type` - Specifies the storage type associated with DB snapshot.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_id` - Provides the VPC ID associated with the DB snapshot.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_db_snapshot_copy` using the snapshot identifier. For example:

```terraform
import {
  to = aws_db_snapshot_copy.example
  id = "my-snapshot"
}
```

Using `terraform import`, import `aws_db_snapshot_copy` using the snapshot identifier. For example:

```console
% terraform import aws_db_snapshot_copy.example my-snapshot
```

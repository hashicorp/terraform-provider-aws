---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_snapshots"
description: |-
  Provides details about an AWS RDS (Relational Database) Snapshots.
---

# Data Source: aws_rds_snapshots

Provides details about an AWS RDS (Relational Database) Snapshots.

## Example Usage

### Basic Usage

```terraform
data "aws_rds_snapshots" "example" {
  db_instance_identifier = "my-db-instance"
}
```

### Filter by snapshot ID

```terraform
data "aws_rds_snapshots" "example" {
  filter {
    name   = "db-snapshot-id"
    values = ["my-snapshot-id"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `db_instance_identifier` - (Optional) Returns the list of snapshots created by the specific db_instance.
* `db_snapshot_identifier` - (Optional) Returns information on a specific snapshot_id.
* `filter` - (Optional) Configuration block(s) used to filter snapshots with AWS supported attributes. Detailed below.
* `include_public` - (Optional) Set this value to `true` to include manual DB snapshots that are public and can be copied or restored by any AWS account, otherwise set this value to `false`. The default is `false`.
* `include_shared` - (Optional) Set this value to `true` to include shared manual DB snapshots from other AWS accounts that this AWS account has been given permission to copy or restore, otherwise set this value to `false`. The default is `false`.
* `snapshot_type` - (Optional) Type of snapshots to be returned. If you don't specify a SnapshotType value, then both automated and manual snapshots are returned. Shared and public DB snapshots are not included in the returned results by default. Possible values are `automated`, `manual`, `shared`, `public` and `awsbackup`.

### `filter` Block

* `name` - (Required) Name of the filter field. Valid values can be found in the RDS DescribeDBSnapshots API Reference.
* `values` - (Required) Set of values accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `snapshots` - List of snapshots.

### `snapshots` Attribute Reference

* `allocated_storage` - Allocated storage size in gigabytes (GB).
* `availability_zone` - Name of the Availability Zone the DB instance was located in at the time of the DB snapshot.
* `db_instance_identifier` - Identifier of the DB instance from which the snapshot was taken.
* `db_snapshot_arn` - ARN for the DB snapshot.
* `db_snapshot_identifier` - Identifier of the DB snapshot.
* `encrypted` - Whether the DB snapshot is encrypted.
* `engine` - Name of the database engine.
* `engine_version` - Version of the database engine.
* `iops` - Provisioned IOPS (I/O operations per second) value of the DB instance at the time of the snapshot.
* `kms_key_id` - ARN for the KMS encryption key.
* `license_model` - License model information for the restored DB instance.
* `option_group_name` - Option group name for the DB snapshot.
* `original_snapshot_create_time` - Time when the snapshot was taken, in Universal Coordinated Time (UTC). Doesn't change when the snapshot is copied.
* `port` - Port that the database engine was listening on at the time of the snapshot.
* `snapshot_create_time` - Time when the snapshot was taken, in Universal Coordinated Time (UTC). Changes when the snapshot is copied.
* `snapshot_type` - Type of the DB snapshot.
* `source_db_snapshot_identifier` - DB snapshot ARN that the DB snapshot was copied from. Only set for cross-account or cross-region copies.
* `source_region` - Region that the DB snapshot was created in or copied from.
* `status` - Status of this DB snapshot.
* `storage_type` - Storage type associated with the DB snapshot.
* `tags` - Map of tags assigned to the snapshot.
* `vpc_id` - ID of the VPC associated with the DB snapshot.

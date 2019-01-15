---
layout: "aws"
page_title: "AWS: aws_docdb_cluster_instance"
sidebar_current: "docs-aws-resource-docdb-cluster-instance"
description: |-
  Provides an DocDB Cluster Resource Instance
---

# aws_docdb_cluster_instance

Provides an DocDB Cluster Resource Instance. A Cluster Instance Resource defines
attributes that are specific to a single instance in a [DocDB Cluster][1].

You do not designate a primary and subsequent replicas. Instead, you simply add DocDB
Instances and DocDB manages the replication. You can use the [count][3]
meta-parameter to make multiple instances and join them all to the same DocDB
Cluster, or you may specify different Cluster Instance resources with various
`instance_class` sizes.

## Example Usage

```hcl
resource "aws_docdb_cluster_instance" "cluster_instances" {
  count              = 2
  identifier         = "docdb-cluster-demo-${count.index}"
  cluster_identifier = "${aws_docdb_cluster.default.id}"
  instance_class     = "db.r4.large"
}

resource "aws_docdb_cluster" "default" {
  cluster_identifier = "docdb-cluster-demo"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  master_username    = "foo"
  master_password    = "barbut8chars"
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/docdb/create-db-instance.html).

The following arguments are supported:

* `identifier` - (Optional, Forces new resource) The indentifier for the DocDB instance, if omitted, Terraform will assign a random, unique identifier.
* `identifier_prefix` - (Optional, Forces new resource) Creates a unique identifier beginning with the specified prefix. Conflicts with `identifer`.
* `cluster_identifier` - (Required) The identifier of the [`aws_docdb_cluster`](/docs/providers/aws/r/docdb_cluster.html) in which to launch this instance.
* `engine` - (Optional) The name of the database engine to be used for the DocDB instance. Defaults to `docdb`. Valid Values: `docdb`.
* `engine_version` - (Optional) The database engine version.
* `instance_class` - (Required) The instance class to use. For details on CPU and memory, see [Scaling for DocDB Instances][2]. DocDB currently
  supports the below instance classes. Please see [AWS Documentation][4] for complete details.
  - db.r4.large
  - db.r4.xlarge
  - db.r4.2xlarge
  - db.r4.4xlarge
  - db.r4.8xlarge
  - db.r4.16xlarge
* `db_subnet_group_name` - (Optional) A DB subnet group to associate with this DB instance. **NOTE:** This must match the `db_subnet_group_name` of the attached [`aws_docdb_cluster`](/docs/providers/aws/r/docdb_cluster.html).
* `apply_immediately` - (Optional) Specifies whether any database modifications
     are applied immediately, or during the next maintenance window. Default is`false`.
* `promotion_tier` - (Optional) Default 0. Failover Priority setting on instance level. The reader who has lower tier has higher priority to get promoter to writer.
* `availability_zone` - (Optional, Computed) The EC2 Availability Zone that the DB instance is created in. See [docs](https://docs.aws.amazon.com/AmazonDocDB/latest/APIReference/API_CreateDBInstance.html) about the details.
* `preferred_maintenance_window` - (Optional) The window to perform maintenance in.
  Syntax: "ddd:hh24:mi-ddd:hh24:mi". Eg: "Mon:00:00-Mon:03:00".
* `auto_minor_version_upgrade` - (Optional) Indicates that minor engine upgrades will be applied automatically to the DB instance during the maintenance window. Default `true`.
* `tags` - (Optional) A mapping of tags to assign to the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of cluster instance
* `availability_zone` - The availability zone of the instance
* `cluster_identifier` - The DocDB Cluster Identifier
* `dbi_resource_id` - The region-unique, immutable identifier for the DB instance.
* `endpoint` - The DNS address for this instance. May not be writable
* `engine_version` - The database engine version
* `engine` - The database engine
* `identifier` - The Instance identifier
* `kms_key_id` - The ARN for the KMS encryption key if one is set to the cluster.
* `port` - The database port
* `preferred_backup_window` - The daily time range during which automated backups are created if automated backups are enabled.
* `status` - The DocDB instance status
* `storage_encrypted` - Specifies whether the DB cluster is encrypted.
* `writer` – Boolean indicating if this instance is writable. `False` indicates this instance is a read replica.

[1]: /docs/providers/aws/r/docdb_cluster.html
[2]: https://docs.aws.amazon.com/documentdb/latest/developerguide/db-cluster-manage-performance.html#db-cluster-manage-scaling-instance
[3]: /docs/configuration/resources.html#count
[4]: https://docs.aws.amazon.com/documentdb/latest/developerguide/db-instance-classes.html#db-instance-class-specs

## Timeouts

`aws_docdb_cluster_instance` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `90 minutes`) Used for Creating Instances, Replicas, and
restoring from Snapshots
- `update` - (Default `90 minutes`) Used for Database modifications
- `delete` - (Default `90 minutes`) Used for destroying databases. This includes
the time required to take snapshots

## Import

DocDB Cluster Instances can be imported using the `identifier`, e.g.

```
$ terraform import aws_docdb_cluster_instance.prod_instance_1 aurora-cluster-instance-1
```

---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_cluster_instance"
description: |-
  Provides an DocumentDB Cluster Resource Instance
---

# Resource: aws_docdb_cluster_instance

Provides an DocumentDB Cluster Resource Instance. A Cluster Instance Resource defines attributes that are specific to a single instance in a [DocumentDB Cluster](/docs/providers/aws/r/docdb_cluster.html).

You do not designate a primary and subsequent replicas. Instead, you simply add DocumentDB Instances and DocumentDB manages the replication. You can use the [count](https://www.terraform.io/docs/configuration/meta-arguments/count.html) meta-parameter to make multiple instances and join them all to the same DocumentDB Cluster, or you may specify different Cluster Instance resources with various `instance_class` sizes.

## Example Usage

```terraform
resource "aws_docdb_cluster_instance" "cluster_instances" {
  count              = 2
  identifier         = "docdb-cluster-demo-${count.index}"
  cluster_identifier = aws_docdb_cluster.default.id
  instance_class     = "db.r5.large"
}

resource "aws_docdb_cluster" "default" {
  cluster_identifier = "docdb-cluster-demo"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  master_username    = "foo"
  master_password    = "barbut8chars"
}
```

## Argument Reference

This resource supports the following arguments:

* `apply_immediately` - (Optional) Whether any database modifications are applied immediately, or during the next maintenance window. Default is`false`.
* `auto_minor_version_upgrade` - (Optional) Parameter does not apply to Amazon DocumentDB. Amazon DocumentDB does not perform minor version upgrades regardless of the value set (see [docs](https://docs.aws.amazon.com/documentdb/latest/developerguide/API_DBInstance.html)). Default `true`.
* `availability_zone` - (Optional, Computed) EC2 Availability Zone that the DB instance is created in. See [docs](https://docs.aws.amazon.com/documentdb/latest/developerguide/API_CreateDBInstance.html) about the details.
* `ca_cert_identifier` - (Optional) Identifier of the certificate authority (CA) certificate for the DB instance.
* `certificate_rotation_restart` – (Optional) Whether to restart the DB instance when rotating its SSL/TLS certificate. By default, AWS restarts the DB instance when you rotate your SSL/TLS certificate. The certificate is not updated until the DB instance is restarted. Set to `"false"` only if you are not using SSL/TLS to connect to the DB instance. If you are using SSL/TLS connections, omit this argument or set to `"true"` to ensure the certificate is properly updated. Valid values: `"true"`, `"false"`, or omit.
* `cluster_identifier` - (Required) Identifier of the [`aws_docdb_cluster`](/docs/providers/aws/r/docdb_cluster.html) in which to launch this instance.
* `copy_tags_to_snapshot` - (Optional, boolean) Copy all DB instance `tags` to snapshots. Default is `false`.
* `enable_performance_insights` - (Optional) Value that indicates whether to enable Performance Insights for the DB Instance. Default `false`. See [docs] (https://docs.aws.amazon.com/documentdb/latest/developerguide/performance-insights.html) about the details.
* `engine` - (Optional) Name of the database engine to be used for the DocumentDB instance. Defaults to `docdb`. Valid Values: `docdb`.
* `identifier_prefix` - (Optional, Forces new resource) Creates a unique identifier beginning with the specified prefix. Conflicts with `identifier`.
* `identifier` - (Optional, Forces new resource) The identifier for the DocumentDB instance, if omitted, Terraform will assign a random, unique identifier.
* `instance_class` - (Required) Instance class to use. For details on CPU and memory, see [Scaling for DocumentDB Instances](https://docs.aws.amazon.com/documentdb/latest/developerguide/db-cluster-manage-performance.html#db-cluster-manage-scaling-instance). See the [`aws_docdb_orderable_db_instance`](/docs/providers/aws/d/docdb_orderable_db_instance.html) data source. See [AWS Documentation](https://docs.aws.amazon.com/documentdb/latest/developerguide/db-instance-classes.html#db-instance-class-specs) for complete details.
* `performance_insights_kms_key_id` - (Optional) KMS key identifier is the key ARN, key ID, alias ARN, or alias name for the KMS key. If you do not specify a value for PerformanceInsightsKMSKeyId, then Amazon DocumentDB uses your default KMS key.
* `preferred_maintenance_window` - (Optional) Window to perform maintenance in. Syntax: "ddd:hh24:mi-ddd:hh24:mi". Eg: "Mon:00:00-Mon:03:00".
* `promotion_tier` - (Optional) Failover Priority setting on instance level. Default `0`. The reader who has lower tier has higher priority to get promoter to writer.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the instance. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of cluster instance
* `db_subnet_group_name` - DB subnet group to associate with this DB instance.
* `dbi_resource_id` - Region-unique, immutable identifier for the DB instance.
* `endpoint` - DNS address for this instance. May not be writable
* `engine_version` - Database engine version
* `kms_key_id` - ARN for the KMS encryption key if one is set to the cluster.
* `port` - Database port
* `preferred_backup_window` - Daily time range during which automated backups are created if automated backups are enabled.
* `storage_encrypted` - Whether the DB cluster is encrypted.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `writer` - Whether this instance is writable. `False` indicates this instance is a read replica.

For more detailed documentation about each argument, refer to the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/docdb/create-db-instance.html).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `90m`)
restoring from Snapshots
- `update` - (Default `90m`)
- `delete` - (Default `90m`)
the time required to take snapshots

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DocumentDB Cluster Instances using the `identifier`. For example:

```terraform
import {
  to = aws_docdb_cluster_instance.prod_instance_1
  id = "aurora-cluster-instance-1"
}
```

Using `terraform import`, import DocumentDB Cluster Instances using the `identifier`. For example:

```console
% terraform import aws_docdb_cluster_instance.prod_instance_1 aurora-cluster-instance-1
```

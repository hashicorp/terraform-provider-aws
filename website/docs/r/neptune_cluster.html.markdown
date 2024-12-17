---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster"
description: |-
  Provides an Neptune Cluster Resource
---

# Resource: aws_neptune_cluster

Provides an Neptune Cluster Resource. A Cluster Resource defines attributes that are
applied to the entire cluster of Neptune Cluster Instances.

Changes to a Neptune Cluster can occur when you manually change a
parameter, such as `backup_retention_period`, and are reflected in the next maintenance
window. Because of this, Terraform may report a difference in its planning
phase because a modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change immediately
(see documentation below).

## Example Usage

```terraform
resource "aws_neptune_cluster" "default" {
  cluster_identifier                  = "neptune-cluster-demo"
  engine                              = "neptune"
  backup_retention_period             = 5
  preferred_backup_window             = "07:00-09:00"
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}
```

~> **Note:** AWS Neptune does not support user name/password–based access control.
See the AWS [Docs](https://docs.aws.amazon.com/neptune/latest/userguide/limits.html) for more information.

## Argument Reference

This resource supports the following arguments:

* `allow_major_version_upgrade` - (Optional) Specifies whether upgrades between different major versions are allowed. You must set it to `true` when providing an `engine_version` parameter that uses a different major version than the DB cluster's current version. Default is `false`.
* `apply_immediately` - (Optional) Specifies whether any cluster modifications are applied immediately, or during the next maintenance window. Default is `false`.
* `availability_zones` - (Optional) A list of EC2 Availability Zones that instances in the Neptune cluster can be created in.
* `backup_retention_period` - (Optional) The days to retain backups for. Default `1`
* `cluster_identifier` - (Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier.
* `cluster_identifier_prefix` - (Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with `cluster_identifier`.
* `copy_tags_to_snapshot` - (Optional) If set to true, tags are copied to any snapshot of the DB cluster that is created.
* `enable_cloudwatch_logs_exports` - (Optional) A list of the log types this DB cluster is configured to export to Cloudwatch Logs. Currently only supports `audit` and `slowquery`.
* `engine` - (Optional) The name of the database engine to be used for this Neptune cluster. Defaults to `neptune`.
* `engine_version` - (Optional) The database engine version.
* `final_snapshot_identifier` - (Optional) The name of your final Neptune snapshot when this Neptune cluster is deleted. If omitted, no final snapshot will be made.
* `global_cluster_identifier` - (Optional) The global cluster identifier specified on [`aws_neptune_global_cluster`](/docs/providers/aws/r/neptune_global_cluster.html).
* `iam_roles` - (Optional) A List of ARNs for the IAM roles to associate to the Neptune Cluster.
* `iam_database_authentication_enabled` - (Optional) Specifies whether or not mappings of AWS Identity and Access Management (IAM) accounts to database accounts is enabled.
* `kms_key_arn` - (Optional) The ARN for the KMS encryption key. When specifying `kms_key_arn`, `storage_encrypted` needs to be set to true.
* `neptune_subnet_group_name` - (Optional) A Neptune subnet group to associate with this Neptune instance.
* `neptune_cluster_parameter_group_name` - (Optional) A cluster parameter group to associate with the cluster.
* `neptune_instance_parameter_group_name` - (Optional) The name of the DB parameter group to apply to all instances of the DB cluster.
* `storage_type` - (Optional) Storage type associated with the cluster `standard/iopt1`. Default: `standard`
* `preferred_backup_window` - (Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter. Time in UTC. Default: A 30-minute window selected at random from an 8-hour block of time per regionE.g., 04:00-09:00
* `preferred_maintenance_window` - (Optional) The weekly time range during which system maintenance can occur, in (UTC) e.g., wed:04:00-wed:04:30
* `port` - (Optional) The port on which the Neptune accepts connections. Default is `8182`.
* `replication_source_identifier` - (Optional) ARN of a source Neptune cluster or Neptune instance if this Neptune cluster is to be created as a Read Replica.
* `skip_final_snapshot` - (Optional) Determines whether a final Neptune snapshot is created before the Neptune cluster is deleted. If true is specified, no Neptune snapshot is created. If false is specified, a Neptune snapshot is created before the Neptune cluster is deleted, using the value from `final_snapshot_identifier`. Default is `false`.
* `snapshot_identifier` - (Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a Neptune cluster snapshot, or the ARN when specifying a Neptune snapshot. Automated snapshots **should not** be used for this attribute, unless from a different cluster. Automated snapshots are deleted as part of cluster destruction when the resource is replaced.
* `storage_encrypted` - (Optional) Specifies whether the Neptune cluster is encrypted. The default is `false` if not specified.
* `tags` - (Optional) A map of tags to assign to the Neptune cluster. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate with the Cluster
* `deletion_protection` - (Optional) A value that indicates whether the DB cluster has deletion protection enabled.The database can't be deleted when deletion protection is enabled. By default, deletion protection is disabled.
* `serverless_v2_scaling_configuration` - (Optional) If set, create the Neptune cluster as a serverless one. See [Serverless](#serverless) for example block attributes.

### Serverless

**Neptune serverless has some limitations. Please see the [limitations on the AWS documentation](https://docs.aws.amazon.com/neptune/latest/userguide/neptune-serverless.html#neptune-serverless-limitations) before jumping into Neptune Serverless.**

Neptune serverless requires that the `engine_version` attribute must be `1.2.0.1` or above. Also, you need to provide a cluster parameter group compatible with the family `neptune1.2`. In the example below, the default cluster parameter group is used.

```terraform
resource "aws_neptune_cluster" "example" {
  cluster_identifier                   = "neptune-cluster-development"
  engine                               = "neptune"
  engine_version                       = "1.2.0.1"
  neptune_cluster_parameter_group_name = "default.neptune1.2"
  skip_final_snapshot                  = true
  apply_immediately                    = true

  serverless_v2_scaling_configuration {}
}

resource "aws_neptune_cluster_instance" "example" {
  cluster_identifier           = aws_neptune_cluster.example.cluster_identifier
  instance_class               = "db.serverless"
  neptune_parameter_group_name = "default.neptune1.2"
}
```

* `min_capacity`: (default: **2.5**) The minimum Neptune Capacity Units (NCUs) for this cluster. Must be greater or equal than **1**. See [AWS Documentation](https://docs.aws.amazon.com/neptune/latest/userguide/neptune-serverless-capacity-scaling.html) for more details.
* `max_capacity`: (default: **128**) The maximum Neptune Capacity Units (NCUs) for this cluster. Must be lower or equal than **128**. See [AWS Documentation](https://docs.aws.amazon.com/neptune/latest/userguide/neptune-serverless-capacity-scaling.html) for more details.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Neptune Cluster Amazon Resource Name (ARN)
* `cluster_resource_id` - The Neptune Cluster Resource ID
* `cluster_members` – List of Neptune Instances that are a part of this cluster
* `endpoint` - The DNS address of the Neptune instance
* `hosted_zone_id` - The Route53 Hosted Zone ID of the endpoint
* `id` - The Neptune Cluster Identifier
* `reader_endpoint` - A read-only endpoint for the Neptune cluster, automatically load-balanced across replicas
* `status` - The Neptune instance status
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_neptune_cluster` using the cluster identifier. For example:

```terraform
import {
  to = aws_neptune_cluster.example
  id = "my-cluster"
}
```

Using `terraform import`, import `aws_neptune_cluster` using the cluster identifier. For example:

```console
% terraform import aws_neptune_cluster.example my-cluster
```

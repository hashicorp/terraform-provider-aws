---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_snapshot"
description: |-
  Provides a MemoryDB Snapshot.
---

# Resource: aws_memorydb_snapshot

Provides a MemoryDB Snapshot.

More information about snapshot and restore can be found in the [MemoryDB User Guide](https://docs.aws.amazon.com/memorydb/latest/devguide/snapshots.html).

## Example Usage

```terraform
resource "aws_memorydb_snapshot" "example" {
  cluster_name = aws_memorydb_cluster.example.name
  name         = "my-snapshot"
}
```

## Argument Reference

This resource supports the following arguments:

* `cluster_name` - (Required, Forces new resource) Name of the MemoryDB cluster to take a snapshot of.
* `name` - (Optional, Forces new resource) Name of the snapshot. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `kms_key_arn` - (Optional, Forces new resource) ARN of the KMS key used to encrypt the snapshot at rest.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the snapshot.
* `arn` - The ARN of the snapshot.
* `cluster_configuration` - The configuration of the cluster from which the snapshot was taken.
    * `description` - Description for the cluster.
    * `engine` - The engine that will run on cluster nodes.
    * `engine_version` - Version number of the Redis engine used by the cluster.
    * `maintenance_window` - The weekly time range during which maintenance on the cluster is performed.
    * `name` - Name of the cluster.
    * `node_type` - Compute and memory capacity of the nodes in the cluster.
    * `num_shards` - Number of shards in the cluster.
    * `parameter_group_name` - Name of the parameter group associated with the cluster.
    * `port` - Port number on which the cluster accepts connections.
    * `snapshot_retention_limit` - Number of days for which MemoryDB retains automatic snapshots before deleting them.
    * `snapshot_window` - The daily time range (in UTC) during which MemoryDB begins taking a daily snapshot of the shard.
    * `subnet_group_name` - Name of the subnet group used by the cluster.
    * `topic_arn` - ARN of the SNS topic to which cluster notifications are sent.
    * `vpc_id` - The VPC in which the cluster exists.
* `source` - Indicates whether the snapshot is from an automatic backup (`automated`) or was created manually (`manual`).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a snapshot using the `name`. For example:

```terraform
import {
  to = aws_memorydb_snapshot.example
  id = "my-snapshot"
}
```

Using `terraform import`, import a snapshot using the `name`. For example:

```console
% terraform import aws_memorydb_snapshot.example my-snapshot
```

---
subcategory: "MemoryDB"
layout: "aws"
page_title: "AWS: aws_memorydb_multi_region_cluster"
description: |-
  Provides a MemoryDB Multi Region Cluster.
---

# Resource: aws_memorydb_multi_region_cluster

Provides a MemoryDB Multi Region Cluster.

More information about MemoryDB can be found in the [Developer Guide](https://docs.aws.amazon.com/memorydb/latest/devguide/what-is-memorydb-for-redis.html).

## Example Usage

```terraform
resource "aws_memorydb_multi_region_cluster" "example" {
  multi_region_cluster_name_suffix = "example"
  node_type                        = "db.r7g.xlarge"
}

resource "aws_memorydb_cluster" "example" {
  acl_name                   = aws_memorydb_acl.example.id
  auto_minor_version_upgrade = false
  name                       = "example"
  node_type                  = "db.t4g.small"
  num_shards                 = 2
  security_group_ids         = [aws_security_group.example.id]
  snapshot_retention_limit   = 7
  subnet_group_name          = aws_memorydb_subnet_group.example.id

  multi_region_cluster_name = aws_memorydb_multi_region_cluster.example.multi_region_cluster_name
}
```

## Argument Reference

The following arguments are required:

* `multi_region_cluster_name_suffix` - (Required, Forces new resource) A suffix to be added to the multi-region cluster name. An AWS generated prefix is automatically applied to the multi-region cluster name when it is created.
* `node_type` - (Required) The node type to be used for the multi-region cluster.

The following arguments are optional:

* `description` - (Optional) description for the multi-region cluster.
* `engine` - (Optional) The name of the engine to be used for the multi-region cluster. Valid values are `redis` and `valkey`.
* `engine_version` - (Optional) The version of the engine to be used for the multi-region cluster. Downgrades are not supported.
* `multi_region_parameter_group_name` - (Optional) The name of the multi-region parameter group to be associated with the cluster.
* `num_shards` - (Optional) The number of shards for the multi-region cluster.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tls_enabled` - (Optional, Forces new resource) A flag to enable in-transit encryption on the cluster.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the multi-region cluster.
* `multi_region_cluster_name` - The name of the multi-region cluster.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a cluster using the `multi_region_cluster_name`. For example:

```terraform
import {
  to = aws_memorydb_multi_region_cluster.example
  id = "virxk-example"
}
```

Using `terraform import`, import a cluster using the `multi_region_cluster_name`. For example:

```console
% terraform import aws_memorydb_multi_region_cluster.example virxk-example
```

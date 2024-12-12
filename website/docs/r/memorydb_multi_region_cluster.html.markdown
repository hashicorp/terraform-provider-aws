---
subcategory: "MemoryDB"
layout: "aws"
page_title: "AWS: aws_memorydb_multi_region_cluster"
description: |-
  Provides a MemoryDB Cluster.
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

* `multi_region_cluster_name_suffix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `node_type` - (Required) The node type to be used for the multi-Region cluster.

The following arguments are optional:

* `description` - (Optional) description for the multi-Region cluster. Defaults to `"Managed by Terraform"`.
* `engine` - (Optional) The name of the engine to be used for the multi-Region cluster. Supported values are `redis` and `valkey`.
* `engine_version` - (Optional) The version of the engine to be used for the multi-Region cluster. Downgrades are not supported.
* `num_shards` - (Optional) The number of shards for the multi-Region cluster.
* `multi_region_parameter_group_name` - (Optional) The name of the multi-Region parameter group to be associated with the cluster.
* `tls_enabled` - (Optional, Forces new resource) A flag to enable in-transit encryption on the cluster. Defaults to `true`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the multi-region cluster.
* `multi_region_cluster_name` - The name of the multi-region cluster.
* `arn` - The ARN of the multi-region cluster.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `120m`)
- `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a cluster using the `name`. For example:

```terraform
import {
  to = aws_memorydb_multi_region_cluster.example
  id = "my-multi-region-cluster"
}
```

Using `terraform import`, import a cluster using the `name`. For example:

```console
% terraform import aws_memorydb_multi_region_cluster.example my-multi-region-cluster
```

---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_node_type"
description: |-
  Information about hardware specifications for an ElastiCache node type.
---

# Data Source: aws_elasticache_node_type

Information about hardware specifications (memory, vCPUs) for an ElastiCache node type.

The ElastiCache API does not expose memory or vCPU information for node types. This data source derives the underlying EC2 instance type from the node type (for example, `cache.t3.medium` corresponds to EC2 instance type `t3.medium`) and looks up its hardware specifications. As a result, node types with no directly corresponding EC2 instance type are not supported.

## Example Usage

```terraform
data "aws_elasticache_node_type" "example" {
  cache_node_type = "cache.t3.medium"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cache_node_type` - (Required) ElastiCache node type, for example `cache.t3.medium`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `burstable_performance_supported` - Whether the node type is a burstable performance instance type.
* `current_generation` - Whether the node type is a current generation instance type.
* `default_cores` - Default number of cores.
* `default_threads_per_core` - Default number of threads per core.
* `default_vcpus` - Default number of vCPUs.
* `ec2_instance_type` - EC2 instance type derived from `cache_node_type` and used to look up hardware specifications.
* `free_tier_eligible` - Whether the node type is eligible for the free tier.
* `id` - ElastiCache node type (same value as `cache_node_type`).
* `memory_size` - Size of memory for the node type, in MiB.
* `supported_architectures` - Supported processor architectures.

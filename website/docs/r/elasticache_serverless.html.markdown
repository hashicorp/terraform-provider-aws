---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_serverless"
description: |-
  Provides an ElastiCache Serverless Cluster resource.
---

# Resource: aws_elasticache_serverless

Provides an ElastiCache Serverlesss Cluster resource which manages memcache or redis.

## Example Usage

### Memcached Serverless

```terraform
resource "aws_elasticache_serverless" "example" {
  engine                = "memcached"
  serverless_cache_name = "example"
  cache_usage_limits {
    data_storage {
      maximum = 10
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 5
    }
  }
  description          = "Test Server"
  kms_key_id           = aws_kms_key.test.arn
  major_engine_version = "1.6"
  security_group_ids   = [aws_security_group.test.id]
  subnet_ids           = aws_subnet.test[*].id
}
```

### Redis Serverless

```terraform
resource "aws_elasticache_serverless" "example" {
  engine                = "redis"
  serverless_cache_name = "example"
  cache_usage_limits {
    data_storage {
      maximum = 10
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 5
    }
  }
  daily_snapshot_time      = "09:00"
  description              = "Test Server"
  kms_key_id               = aws_kms_key.test.arn
  major_engine_version     = "1.6"
  snapshot_retention_limit = 1
  security_group_ids       = [aws_security_group.test.id]
  subnet_ids               = aws_subnet.test[*].id
}
```

## Argument Reference

The following arguments are required:

* `engine` – (Required) Name of the cache engine to be used for this cache cluster. Valid values are `memcached` or `redis`.
* `serverless_cache_name` – (Required) The Cluster name which serves as a unique identifier to the serverless cache

The following arguments are optional:

* `cache_usage_limits` - (Optional) Sets the cache usage limits for storage and ElastiCache Processing Units for the cache. See configuration below.
* `daily_snapshott_time` - (Optional) The daily time that snapshots will be created from the new serverless cache. Only supported for engine type `"redis"`. Defaults to `0`.
* `description` - (Optional) User-provided description for the serverless cache. The default is NULL.
* `kms_key_id` - (Optional) ARN of the customer managed key for encrypting the data at rest. If no KMS key is provided, a default service key is used.
* `major_engine_version` – (Optional) The version of the cache engine that will be used to create the serverless cache.
  See [Describe Cache Engine Versions](https://docs.aws.amazon.com/cli/latest/reference/elasticache/describe-cache-engine-versions.html) in the AWS Documentation for supported versions.
* `security_group_ids` - (Optional) A list of the one or more VPC security groups to be associated with the serverless cache. The security group will authorize traffic access for the VPC end-point (private-link). If no other information is given this will be the VPC’s Default Security Group that is associated with the cluster VPC end-point.
* `snapshot_arns` - (Optional, Redis only) The list of ARN(s) of the snapshot that the new serverless cache will be created from. Available for Redis only.
* `snapshot_retention_limit` - (Optional, Redis only) The number of snapshots that will be retained for the serverless cache that is being created. As new snapshots beyond this limit are added, the oldest snapshots will be deleted on a rolling basis. Available for Redis only.
* `subnet_ids` – (Optional) A list of the identifiers of the subnets where the VPC endpoint for the serverless cache will be deployed. All the subnetIds must belong to the same VPC.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_group_id` - (Optional) The identifier of the UserGroup to be associated with the serverless cache. Available for Redis only. Default is NULL.

### CacheUsageLimits Configuration

* `data_storage` - The maximum data storage limit in the cache, expressed in Gigabytes. See Data Storage config for more details.
* `ecpu_per_second` - The configuration for the number of ElastiCache Processing Units (ECPU) the cache can consume per second.See config block for more details.

### DataStorage Configuration

* `maximum` - The upper limit for data storage the cache is set to use. Set as Integer.
* `unit` - The unit that the storage is measured in, in GB.

### ECPUPerSecond Configuration

* `maximum` - The upper limit for data storage the cache is set to use. Set as Integer.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the serverless cache.
* `create_time` - Timestamp of when the serverless cache was created.
* `endpoint` - Represents the information required for client programs to connect to a cache node. See config below for details.
* `full_engine_version` - The name and version number of the engine the serverless cache is compatible with.
* `major_engine_version` - The version number of the engine the serverless cache is compatible with.
* `reader_endpoint` - Represents the information required for client programs to connect to a cache node. See config below for details.
* `status` - The current status of the serverless cache. The allowed values are CREATING, AVAILABLE, DELETING, CREATE-FAILED and MODIFYING.

### Endpoint Configuration

* `address` - The DNS hostname of the cache node.
* `port` - The port number that the cache engine is listening on. Set as integer.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `40m`)
- `update` - (Default `80m`)
- `delete` - (Default `40m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache Serverlesss Clusters using the `serverless_cache_name`. For example:

```terraform
import {
  to                    = aws_elasticache_serverless.my_cluster
  serverless_cache_name = "my_cluster"
}
```

Using `terraform import`, import ElastiCache Serverless Clusters using the `sserverless_cache_name`. For example:

```console
% terraform import aws_elasticache_serverless.my_cluster my_cluster
```

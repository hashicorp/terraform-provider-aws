---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_serverless_cache"
description: |-
  Get information on an ElastiCache Serverless Cache resource.
---

# Data Source: aws_elasticache_serverless_cache

Use this data source to get information about an ElastiCache Serverless Cache.

## Example Usage

```terraform
data "aws_elasticache_serverless_cache" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` – (Required) Identifier for the serverless cache.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the serverless cache.
* `cache_usage_limits` - The cache usage limits for storage and ElastiCache Processing Units for the cache. See [`cache_usage_limits` Block](#cache_usage_limits-block) for details.
* `create_time` - Timestamp of when the serverless cache was created.
* `daily_snapshot_time` - The daily time that snapshots will be created from the new serverless cache. Only available for engine type `"redis"`.
* `description` - Description of the serverless cache.
* `endpoint` - Represents the information required for client programs to connect to the cache. See [`endpoint` Block](#endpoint-block) for details.
* `engine` – Name of the cache engine.
* `full_engine_version` - The name and version number of the engine the serverless cache is compatible with.
* `kms_key_id` - ARN of the customer managed key for encrypting the data at rest.
* `major_engine_version` - The version number of the engine the serverless cache is compatible with.
* `reader_endpoint` - Represents the information required for client programs to connect to a cache node. See [`reader_endpoint` Block](#reader_endpoint-block) for details.
* `security_group_ids` - A list of the one or more VPC security groups associated with the serverless cache.
* `snapshot_retention_limit` - The number of snapshots that will be retained for the serverless cache. Available for Redis only.
* `status` - The current status of the serverless cache.
* `subnet_ids` – A list of the identifiers of the subnets where the VPC endpoint for the serverless cache are deployed.
* `user_group_id` - The identifier of the UserGroup associated with the serverless cache. Available for Redis only.

### `cache_usage_limits` Block

The `cache_usage_limits` block supports the following attributes:

* `data_storage` - The maximum data storage limit in the cache, expressed in Gigabytes. See [`data_storage` Block](#data_storage-block) for details.
* `ecpu_per_second` - The configured number of ElastiCache Processing Units (ECPU) the cache can consume per second. See [`ecpu_per_second` Block](#ecpu_per_second-block) for details.

### `data_storage` Block

The `data_storage` block supports the following attributes:

* `minimum` - The lower limit for data storage the cache is set to use.
* `maximum` - The upper limit for data storage the cache is set to use.
* `unit` - The unit that the storage is measured in.

### `ecpu_per_second` Block

The `ecpu_per_second` block supports the following attributes:

* `minimum` - The minimum number of ECPUs the cache can consume per second.
* `maximum` - The maximum number of ECPUs the cache can consume per second.

### `endpoint` Block

The `endpoint` block exports the following attributes:

* `address` - The DNS hostname of the cache node.
* `port` - The port number that the cache engine is listening on. Set as integer.

### `reader_endpoint` Block

The `reader_endpoint` block exports the following attributes:

* `address` - The DNS hostname of the cache node.
* `port` - The port number that the cache engine is listening on. Set as integer.

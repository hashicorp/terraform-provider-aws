---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_reserved_cache_node_offering"
description: |-
  Information about a single ElastiCache Reserved Cache Node Offering.
---

# Data Source: aws_elasticache_reserved_cache_node_offering

Information about a single ElastiCache Reserved Cache Node Offering.

## Example Usage

```terraform
data "aws_elasticache_reserved_cache_node_offering" "example" {
  cache_node_type     = "cache.t4g.small"
  duration            = "P1Y"
  offering_type       = "No Upfront"
  product_description = "redis"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `cache_node_type` - (Required) Node type for the reserved cache node.
  See AWS documentation for information on [supported node types for Redis](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html) and [guidance on selecting node types for Redis](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/nodes-select-size.html).
  See AWS documentation for information on [supported node types for Memcached](https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/CacheNodes.SupportedTypes.html) and [guidance on selecting node types for Memcached](https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/nodes-select-size.html).
* `duration` - (Required) Duration of the reservation in RFC3339 duration format.
  Valid values are `P1Y` (one year) and `P3Y` (three years).
* `offering_type` - (Required) Offering type of this reserved cache node.
  For the latest generation of nodes (e.g. M5, R5, T4 and newer) valid values are `No Upfront`, `Partial Upfront`, and `All Upfront`.
  For other current generation nodes (i.e. T2, M3, M4, R3, or R4) the only valid value is `Heavy Utilization`.
  For previous generation modes (i.e. T1, M1, M2, or C1) valid values are `Heavy Utilization`, `Medium Utilization`, and `Light Utilization`.
* `product_description` - (Required) Engine type for the reserved cache node.
  Valid values are `redis`, `valkey` and `memcached`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the reservation. Same as `offering_id`.
* `fixed_price` - Fixed price charged for this reserved cache node.
* `offering_id` - Unique identifier for the reservation.

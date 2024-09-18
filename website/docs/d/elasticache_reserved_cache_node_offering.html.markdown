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

The following arguments are supported:

* `cache_node_type` - (Required) Node type for the reserved cache node.
* `duration` - (Required) Duration of the reservation in RFC3339 duration format.
  Valid values are `P1Y` (one year) and `P3Y` (three years).
* `offering_type` - (Required) Offering type of this reserved cache node.
* `product_description` - (Required) Description of the reserved cache node.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Unique identifier for the reservation. Same as `offering_id`.
* `fixed_price` - Fixed price charged for this reserved cache node.
* `offering_id` - Unique identifier for the reservation.

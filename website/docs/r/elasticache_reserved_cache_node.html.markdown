---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_reserved_cache_node"
description: |-
  Manages an ElastiCache Reserved Cache Node
---

# Resource: aws_elasticache_reserved_cache_node

Manages an ElastiCache Reserved Cache Node.

~> **NOTE:** Once created, a reservation is valid for the `duration` of the provided `offering_id` and cannot be deleted. Performing a `destroy` will only remove the resource from state. For more information see [ElastiCache Reserved Nodes Documentation](https://aws.amazon.com/elasticache/reserved-cache-nodes/) and [PurchaseReservedCacheNodesOffering](https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_PurchaseReservedCacheNodesOffering.html).

~> **NOTE:** Due to the expense of testing this resource, we provide it as best effort. If you find it useful, and have the ability to help test or notice issues, consider reaching out to us on [GitHub](https://github.com/hashicorp/terraform-provider-aws).

## Example Usage

```terraform
data "aws_elasticache_reserved_cache_node_offering" "test" {
  cache_node_type     = "cache.t4g.small"
  duration            = "P1Y"
  offering_type       = "No Upfront"
  product_description = "redis"
}

resource "aws_elasticache_reserved_cache_node" "example" {
  offering_id      = data.aws_elasticache_reserved_cache_node_offering.test.offering_id
  reservation_id   = "optionalCustomReservationID"
  cache_node_count = 3
}
```

## Argument Reference

The following arguments are required:

* `offering_id` - (Required) ID of the reserved cache node offering to purchase. To determine an `offering_id`, see the `aws_elasticache_reserved_cache_node_offering` data source.

The following arguments are optional:

* `instance_count` - (Optional) Number of cache node instances to reserve. Default value is `1`.
* `reservation_id` - (Optional) Customer-specified identifier to track this reservation.
* `tags` - (Optional) Map of tags to assign to the reservation. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN for the reserved cache node.
* `id` - Unique identifier for the reservation. same as `reservation_id`.
* `duration` - Duration of the reservation as an RFC3339 duration.
* `fixed_price` – Fixed price charged for this reserved cache node.
* `cache_node_type` - Cache node type for the reserved cache nodes.
* `offering_type` - Offering type of this reserved cache node.
* `product_description` - Description of the reserved cache node.
* `recurring_charges` - Recurring price charged to run this reserved cache node.
* `start_time` - Time the reservation started.
* `state` - State of the reserved cache node.
* `usage_price` - Hourly price charged for this reserved cache node.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `10m`)
- `delete` - (Default `1m`)

## Import

RDS DB Instance Reservations can be imported using the `instance_id`, e.g.,

```
$ terraform import aws_elasticache_reserved_cache_node.reservation_node CustomReservationID
```

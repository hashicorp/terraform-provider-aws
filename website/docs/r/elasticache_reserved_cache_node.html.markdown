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
data "aws_elasticache_reserved_cache_node_offering" "example" {
  cache_node_type     = "cache.t4g.small"
  duration            = "P1Y"
  offering_type       = "No Upfront"
  product_description = "redis"
}

resource "aws_elasticache_reserved_cache_node" "example" {
  reserved_cache_nodes_offering_id = data.aws_elasticache_reserved_cache_node_offering.example.offering_id
  id                               = "optionalCustomReservationID"
  cache_node_count                 = 3
}
```

## Argument Reference

The following arguments are required:

* `reserved_cache_nodes_offering_id` - (Required) ID of the reserved cache node offering to purchase.
  To determine an `reserved_cache_nodes_offering_id`, see the `aws_elasticache_reserved_cache_node_offering` data source.

The following arguments are optional:

* `cache_node_count` - (Optional) Number of cache node instances to reserve.
  Default value is `1`.
* `id` - (Optional) Customer-specified identifier to track this reservation.
  If not specified, AWS will assign a random ID.
* `tags` - (Optional) Map of tags to assign to the reservation. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN for the reserved cache node.
* `duration` - Duration of the reservation as an RFC3339 duration.
* `fixed_price` – Fixed price charged for this reserved cache node.
* `cache_node_type` - Node type for the reserved cache nodes.
* `offering_type` - Offering type of this reserved cache node.
* `product_description` - Engine type for the reserved cache node.
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

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ElastiCache Reserved Cache Nodes using the `id`. For example:

```terraform
import {
  to = aws_elasticache_reserved_cache_node.example
  id = "CustomReservationID"
}
```

Using `terraform import`, import ElastiCache Reserved Cache Node using the `id`. For example:

```console
% terraform import aws_elasticache_reserved_cache_node.example CustomReservationID
```

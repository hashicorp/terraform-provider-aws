---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_public_ipv4_pools"
description: |-
  Terraform data source for managing AWS VPC (Virtual Private Cloud) Public IPv4 Pools.
---

# Data Source: aws_ec2_public_ipv4_pools

Terraform data source for managing an AWS VPC (Virtual Private Cloud) Public IPv4 Pool

## Example Usage

### Basic Usage

```terraform
data "aws_vpc_public_ipv4_pool" "example" {
  pool_ids = "ipv4pool-ec2-000df99cff0c1ec10"
}
```

### Usage with Filter
```terraform
data "aws_vpc_public_ipv4_pool" "example" {
  filter {
    name   = "tag-key"
    values = ["ExampleTagKey"]
  }
}
```

## Argument Reference

The following arguments are required:

* `pool_id` - (Required) AWS resource IDs of a public IPv4 pool (as a string) for which this data source will fetch detailed information.

The following arguments are optional:

* `filter` - (Optional) One or more filters for results. Supported filters include `tag` and `tag-key`.
* `tags` - (Optional) One or more tags, which are used to filter results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `pool` - Record containing information about a Public IPv4 Pool. Contents:
  - `description` - Description of the pool, if any.
  - `network_border_group` - Name of the location from which the address pool is advertised.
  - `pool_address_ranges` - List of Address Ranges in the Pool; each address range record contains:
    - `address_count` - Number of addresses in the range.
    - `available_address_count` - Number of available addresses in the range.
    - `first_address` - First address in the range.
    - `last_address` - Last address in the range.
  - `tags` - Any tags for the address pool.
  - `total_address_count` - Total number of addresses in the pool.
  - `total_available_address_count` - Total number of available addresses in the pool.

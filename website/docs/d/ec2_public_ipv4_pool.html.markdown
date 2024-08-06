---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_public_ipv4_pool"
description: |-
  Provides details about a specific AWS EC2 Public IPv4 Pool.
---

# Data Source: aws_ec2_public_ipv4_pool

Provides details about a specific AWS EC2 Public IPv4 Pool.

## Example Usage

### Basic Usage

```terraform
data "aws_ec2_public_ipv4_pool" "example" {
  pool_id = "ipv4pool-ec2-000df99cff0c1ec10"
}
```

## Argument Reference

The following arguments are required:

* `pool_id` - (Required) AWS resource IDs of a public IPv4 pool (as a string) for which this data source will fetch detailed information.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the pool, if any.
* `network_border_group` - Name of the location from which the address pool is advertised.
* pool_address_ranges` - List of Address Ranges in the Pool; each address range record contains:
    * `address_count` - Number of addresses in the range.
    * `available_address_count` - Number of available addresses in the range.
    * `first_address` - First address in the range.
    * `last_address` - Last address in the range.
* `tags` - Any tags for the address pool.
* `total_address_count` - Total number of addresses in the pool.
* `total_available_address_count` - Total number of available addresses in the pool.

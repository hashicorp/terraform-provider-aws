---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool"
description: |-
    Returns details about the first IPAM pool that matches search parameters provided.
---

# Data Source: aws_vpc_ipam_pool

`aws_vpc_ipam_pool` provides details about an IPAM pool.

This resource can prove useful when an ipam pool was created in another root
module and you need the pool's id as an input variable. For example, pools
can be shared via RAM and used to create vpcs with CIDRs from that pool.

## Example Usage

The following example shows an account that has only 1 pool, perhaps shared
via RAM, and using that pool id to create a VPC with a CIDR derived from
AWS IPAM.

```terraform
data "aws_vpc_ipam_pool" "test" {}

resource "aws_vpc" "test" {
  ipv4_ipam_pool_id   = data.aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = 28
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
IPAM pools in the current region. The given filters must match exactly one
IPAM pool whose data will be exported as attributes.

* `id` - (Optional) The ID of the specific IPAM pool to retrieve.
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeIpamPools.html).

* `values` - (Required) Set of values that are accepted for the given field.
  An IPAM pool will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected IPAM pool.

The following attribute is additionally exported:

* `address_family` - The IP protocol assigned to this pool.
* `publicly_advertisable` - Defines whether or not IPv6 pool space is publicly ∂advertisable over the internet.
* `allocation_default_netmask_length` - A default netmask length for allocations added to this pool. If, for example, the CIDR assigned to this pool is 10.0.0.0/8 and you enter 16 here, new allocations will default to 10.0.0.0/16.
* `allocation_max_netmask_length` - The maximum netmask length that will be required for CIDR allocations in this pool.
* `allocation_min_netmask_length` - The minimum netmask length that will be required for CIDR allocations in this pool.
* `allocation_resource_tags` - Tags that are required to create resources in using this pool.
* `arn` - Amazon Resource Name (ARN) of the pool
* `auto_import` - If enabled, IPAM will continuously look for resources within the CIDR range of this pool and automatically import them as allocations into your IPAM.
* `aws_service` - Limits which service in AWS that the pool can be used in. "ec2", for example, allows users to use space for Elastic IP addresses and VPCs.
* `description` - A description for the IPAM pool.
* `ipam_scope_id` - The ID of the scope the pool belongs to.
* `locale` - Locale is the Region where your pool is available for allocations. You can only create pools with locales that match the operating Regions of the IPAM. You can only create VPCs from a pool whose locale matches the VPC's Region.
* `ipam_pool_id` - The ID of the IPAM pool.
* `source_ipam_pool_id` - The ID of the source IPAM pool.
* `tags` - A map of tags to assigned to the resource.

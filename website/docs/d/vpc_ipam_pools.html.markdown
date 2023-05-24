---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pools"
description: |-
    Returns details about IPAM pools that match the search parameters provided.
---

# Data Source: aws_vpc_ipam_pools

`aws_vpc_ipam_pools` provides details about IPAM pools.

This resource can prove useful when IPAM pools are created in another root
module and you need the pool ids as input variables. For example, pools
can be shared via RAM and used to create vpcs with CIDRs from that pool.

## Example Usage

```terraform
data "aws_vpc_ipam_pools" "test" {
  filter {
    name   = "description"
    values = ["*test*"]
  }

  filter {
    name   = "address-family"
    values = ["ipv4"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
IPAM Pools in the current region.

* `filter` - (Required) Custom filter block as described below.

### filter

* `name` - (Required) The name of the filter. Filter names are case-sensitive.
* `values` - (Required) The filter values. Filter values are case-sensitive.

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `ipam_pools` - List of IPAM pools and their attributes. See below for details

### ipam_pools

The following attributes are available on each pool entry found.

* `address_family` - IP protocol assigned to this pool.
* `allocation_default_netmask_length` - A default netmask length for allocations added to this pool. If, for example, the CIDR assigned to this pool is `10.0.0.0/8` and you enter 16 here, new allocations will default to `10.0.0.0/16`.
* `allocation_max_netmask_length` - The maximum netmask length that will be required for CIDR allocations in this pool.
* `allocation_min_netmask_length` - The minimum netmask length that will be required for CIDR allocations in this pool.
* `allocation_resource_tags` - Tags that are required to create resources in using this pool.
* `arn` - ARN of the pool
* `auto_import` - If enabled, IPAM will continuously look for resources within the CIDR range of this pool and automatically import them as allocations into your IPAM.
* `aws_service` - Limits which service in AWS that the pool can be used in. `ec2` for example, allows users to use space for Elastic IP addresses and VPCs.
* `description` - Description for the IPAM pool.
* `id` - ID of the IPAM pool.
* `ipam_scope_id` - ID of the scope the pool belongs to.
* `locale` - Locale is the Region where your pool is available for allocations. You can only create pools with locales that match the operating Regions of the IPAM. You can only create VPCs from a pool whose locale matches the VPC's Region.
* `publicly_advertisable` - Defines whether or not IPv6 pool space is publicly advertisable over the internet.
* `source_ipam_pool_id` - ID of the source IPAM pool.
* `tags` - Map of tags to assigned to the resource.

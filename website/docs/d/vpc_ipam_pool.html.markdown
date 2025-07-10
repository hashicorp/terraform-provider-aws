---
subcategory: "VPC IPAM (IP Address Manager)"
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
data "aws_vpc_ipam_pool" "test" {
  filter {
    name   = "description"
    values = ["*test*"]
  }

  filter {
    name   = "address-family"
    values = ["ipv4"]
  }
}

resource "aws_vpc" "test" {
  ipv4_ipam_pool_id   = data.aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = 28
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `ipam_pool_id` - (Optional) ID of the IPAM pool you would like information on.
* `filter` - (Optional) Custom filter block as described below.

### `filter`

* `name` - (Required) The name of the filter. Filter names are case-sensitive.
* `values` - (Required) The filter values. Filter values are case-sensitive.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

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

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

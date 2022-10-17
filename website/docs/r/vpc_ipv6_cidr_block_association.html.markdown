---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_ipv6_cidr_block_association"
description: |-
  Associate additional IPv6 CIDR blocks with a VPC
---

# Resource: aws_vpc_ipv6_cidr_block_association

Provides a resource to associate additional IPv6 CIDR blocks with a VPC.

The `aws_vpc_ipv6_cidr_block_association` resource allows IPv6 CIDR blocks to be added to the VPC.

## Example Usage

```terraform
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv6_cidr_block_association" "test" {
  assign_generated_ipv6_cidr_block = true
  vpc_id                           = aws_vpc.test.id
}
```

## Argument Reference

The following arguments are supported:

- `assign_generated_ipv6_cidr_block` - (Optional) Requests an Amazon-provided IPv6 CIDR block with a /56 prefix length for the VPC. You cannot specify the range of IPv6 addresses, or the size of the CIDR block. Default is `false`. Conflicts with `ipv6_pam_pool_id`, `ipv6_pool`, `ipv6_cidr_block` and `ipv6_netmask_length`.
- `ipv6_cidr_block` - (Optional) The IPv6 CIDR block for the VPC. CIDR can be explicitly set or it can be derived from IPAM using `ipv6_netmask_length`. This parameter is required if `ipv6_netmask_length` is not set and he IPAM pool does not have `allocation_default_netmask` set. Conflicts with `assign_generated_ipv6_cidr_block`.
- `ipv6_ipam_pool_id` - (Optional) The ID of an IPv6 IPAM pool you want to use for allocating this VPC's CIDR. IPAM is a VPC feature that you can use to automate your IP address management workflows including assigning, tracking, troubleshooting, and auditing IP addresses across AWS Regions and accounts. Conflict with `assign_generated_ipv6_cidr_block` and `ipv6_ipam_pool_id`.
- `ipv6_netmask_length` - (Optional) The netmask length of the IPv6 CIDR you want to allocate to this VPC. Requires specifying a `ipv6_ipam_pool_id`. This parameter is optional if the IPAM pool has `allocation_default_netmask` set, otherwise it or `cidr_block` are required. Conflicts with `assign_generated_ipv6_cidr_block` and `ipv6_ipam_pool_id`.
- `ipv6_pool` - (Optional) The  ID of an IPv6 address pool from which to allocate the IPv6 CIDR block. Conflicts with `ipv6_pam_pool_id`, `ipv6_pool`.
- `vpc_id` - (Required) The ID of the VPC to make the association with.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the VPC CIDR association
- `ipv6_cidr_block` - The assigned IPv6 CIDR blok

## Import

`aws_vpc_ipv6_cidr_block_association` can be imported by using the VPC CIDR Association ID, e.g.,

```
$ terraform import aws_vpc_ipv6_cidr_block_association.example vpc-cidr-assoc-xxxxxxxx
```

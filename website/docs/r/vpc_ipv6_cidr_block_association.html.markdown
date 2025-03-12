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
  ipv6_ipam_pool_id = aws_vpc_ipam_pool.test.id
  vpc_id            = aws_vpc.test.id
}
```

## Argument Reference

This resource supports the following arguments:

* `assign_generated_ipv6_cidr_block` - (Optional) Requests an Amazon-provided IPv6 CIDR block with a /56 prefix length for the VPC. You cannot specify the range of IPv6 addresses, or the size of the CIDR block. Default is `false`. Conflicts with `ipv6_pam_pool_id`, `ipv6_pool`, `ipv6_cidr_block` and `ipv6_netmask_length`.
* `ipv6_cidr_block` - (Optional) The IPv6 CIDR block for the VPC. CIDR can be explicitly set or it can be derived from IPAM using `ipv6_netmask_length`. This parameter is required if `ipv6_netmask_length` is not set and the IPAM pool does not have `allocation_default_netmask` set. Conflicts with `assign_generated_ipv6_cidr_block`.
* `ipv6_ipam_pool_id` - (Optional) The ID of an IPv6 IPAM pool you want to use for allocating this VPC's CIDR. IPAM is a VPC feature that you can use to automate your IP address management workflows including assigning, tracking, troubleshooting, and auditing IP addresses across AWS Regions and accounts. Conflict with `assign_generated_ipv6_cidr_block` and `ipv6_pool`.
* `ipv6_netmask_length` - (Optional) The netmask length of the IPv6 CIDR you want to allocate to this VPC. Requires specifying a `ipv6_ipam_pool_id`. This parameter is optional if the IPAM pool has `allocation_default_netmask` set, otherwise it or `ipv6_cidr_block` are required. Conflicts with `ipv6_cidr_block`.
* `ipv6_pool` - (Optional) The  ID of an IPv6 address pool from which to allocate the IPv6 CIDR block. Conflicts with `assign_generated_ipv6_cidr_block` and `ipv6_pam_pool_id`.
* `vpc_id` - (Required) The ID of the VPC to make the association with.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC CIDR association.
* `ip_source` - The source that allocated the IP address space. Values: `amazon`, `byoip`, `none`.
* `ipv6_address_attribute` - Public IPv6 addresses are those advertised on the internet from AWS. Private IP addresses are not and cannot be advertised on the internet from AWS. Values: `public`, `private`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_vpc_ipv6_cidr_block_association` using the VPC CIDR association ID and optionally the IPv6 IPAM pool ID and netmask length. For example:

```terraform
import {
  to = aws_vpc_ipv6_cidr_block_association.example
  id = "vpc-cidr-assoc-0754129087e149dcd"
}
```

or

```terraform
import {
  to = aws_vpc_ipv6_cidr_block_association.example
  id = "vpc-cidr-assoc-0754129087e149dcd,ipam-pool-0611d1d6bbc05ce60"
}
```

or

```terraform
import {
  to = aws_vpc_ipv6_cidr_block_association.example
  id = "vpc-cidr-assoc-0754129087e149dcd,ipam-pool-0611d1d6bbc05ce60,56"
}
```

Using `terraform import`, import `aws_vpc_ipv6_cidr_block_association` using the VPC CIDR association ID and optionally the IPv6 IPAM pool ID and netmask length. For example:

```console
% terraform import aws_vpc_ipv6_cidr_block_association.example vpc-cidr-assoc-0754129087e149dcd
```

or

```console
% terraform import aws_vpc_ipv6_cidr_block_association.example vpc-cidr-assoc-0754129087e149dcd,ipam-pool-0611d1d6bbc05ce60
```

or

```console
% terraform import aws_vpc_ipv6_cidr_block_association.example vpc-cidr-assoc-0754129087e149dcd,ipam-pool-0611d1d6bbc05ce60,56
```

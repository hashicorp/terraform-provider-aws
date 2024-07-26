---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool"
description: |-
  Provides a IP address pool resource for IPAM.
---

# Resource: aws_vpc_ipam_pool

Provides an IP address pool resource for IPAM.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "example" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "example" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.example.private_default_scope_id
  locale         = data.aws_region.current.name
}
```

Nested Pools:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "example" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}

resource "aws_vpc_ipam_pool" "parent" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.example.private_default_scope_id
}

resource "aws_vpc_ipam_pool_cidr" "parent_test" {
  ipam_pool_id = aws_vpc_ipam_pool.parent.id
  cidr         = "172.20.0.0/16"
}

resource "aws_vpc_ipam_pool" "child" {
  address_family      = "ipv4"
  ipam_scope_id       = aws_vpc_ipam.example.private_default_scope_id
  locale              = data.aws_region.current.name
  source_ipam_pool_id = aws_vpc_ipam_pool.parent.id
}


resource "aws_vpc_ipam_pool_cidr" "child_test" {
  ipam_pool_id = aws_vpc_ipam_pool.child.id
  cidr         = "172.20.0.0/24"
}
```

## Argument Reference

This resource supports the following arguments:

* `address_family` - (Required) The IP protocol assigned to this pool. You must choose either IPv4 or IPv6 protocol for a pool.
* `allocation_default_netmask_length` - (Optional) A default netmask length for allocations added to this pool. If, for example, the CIDR assigned to this pool is 10.0.0.0/8 and you enter 16 here, new allocations will default to 10.0.0.0/16 (unless you provide a different netmask value when you create the new allocation).
* `allocation_max_netmask_length` - (Optional) The maximum netmask length that will be required for CIDR allocations in this pool.
* `allocation_min_netmask_length` - (Optional) The minimum netmask length that will be required for CIDR allocations in this pool.
* `allocation_resource_tags` - (Optional) Tags that are required for resources that use CIDRs from this IPAM pool. Resources that do not have these tags will not be allowed to allocate space from the pool. If the resources have their tags changed after they have allocated space or if the allocation tagging requirements are changed on the pool, the resource may be marked as noncompliant.
* `auto_import` - (Optional) If you include this argument, IPAM automatically imports any VPCs you have in your scope that fall
within the CIDR range in the pool.
* `aws_service` - (Optional) Limits which AWS service the pool can be used in. Only useable on public scopes. Valid Values: `ec2`.
* `cascade` - (Optional) Enables you to quickly delete an IPAM pool and all resources within that pool, including provisioned CIDRs, allocations, and other pools.
* `description` - (Optional) A description for the IPAM pool.
* `ipam_scope_id` - (Required) The ID of the scope in which you would like to create the IPAM pool.
* `locale` - (Optional) The locale in which you would like to create the IPAM pool. Locale is the Region where you want to make an IPAM pool available for allocations. You can only create pools with locales that match the operating Regions of the IPAM. You can only create VPCs from a pool whose locale matches the VPC's Region. Possible values: Any AWS region, such as `us-east-1`.
* `publicly_advertisable` - (Optional) Defines whether or not IPv6 pool space is publicly advertisable over the internet. This argument is required if `address_family = "ipv6"` and `public_ip_source = "byoip"`, default is `false`. This option is not available for IPv4 pool space or if `public_ip_source = "amazon"`.
* `public_ip_source` - (Optional) The IP address source for pools in the public scope. Only used for provisioning IP address CIDRs to pools in the public scope. Valid values are `byoip` or `amazon`. Default is `byoip`.
* `source_ipam_pool_id` - (Optional) The ID of the source IPAM pool. Use this argument to create a child pool within an existing pool.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of IPAM
* `id` - The ID of the IPAM
* `state` - The ID of the IPAM
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the IPAM pool `id`. For example:

```terraform
import {
  to = aws_vpc_ipam_pool.example
  id = "ipam-pool-0958f95207d978e1e"
}
```

Using `terraform import`, import IPAMs using the IPAM pool `id`. For example:

```console
% terraform import aws_vpc_ipam_pool.example ipam-pool-0958f95207d978e1e
```

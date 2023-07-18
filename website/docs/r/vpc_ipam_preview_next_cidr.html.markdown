---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_preview_next_cidr"
description: |-
  Previews a CIDR from an IPAM address pool.
---

# Resource: aws_vpc_ipam_preview_next_cidr

Previews a CIDR from an IPAM address pool. Only works for private IPv4.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam_preview_next_cidr" "example" {
  ipam_pool_id   = aws_vpc_ipam_pool.example.id
  netmask_length = 28

  disallowed_cidrs = [
    "172.2.0.0/32"
  ]

  depends_on = [
    aws_vpc_ipam_pool_cidr.example
  ]
}

resource "aws_vpc_ipam_pool_cidr" "example" {
  ipam_pool_id = aws_vpc_ipam_pool.example.id
  cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool" "example" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.example.private_default_scope_id
  locale         = data.aws_region.current.name
}

resource "aws_vpc_ipam" "example" {
  operating_regions {
    region_name = data.aws_region.current.name
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `disallowed_cidrs` - (Optional) Exclude a particular CIDR range from being returned by the pool.
* `ipam_pool_id` - (Required) The ID of the pool to which you want to assign a CIDR.
* `netmask_length` - (Optional) The netmask length of the CIDR you would like to preview from the IPAM pool.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `cidr` - The previewed CIDR from the pool.
* `id` - The ID of the preview.

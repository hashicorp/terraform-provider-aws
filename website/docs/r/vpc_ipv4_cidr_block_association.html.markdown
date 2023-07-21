---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_ipv4_cidr_block_association"
description: |-
  Associate additional IPv4 CIDR blocks with a VPC
---

# Resource: aws_vpc_ipv4_cidr_block_association

Provides a resource to associate additional IPv4 CIDR blocks with a VPC.

When a VPC is created, a primary IPv4 CIDR block for the VPC must be specified.
The `aws_vpc_ipv4_cidr_block_association` resource allows further IPv4 CIDR blocks to be added to the VPC.

## Example Usage

```terraform
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "172.2.0.0/16"
}
```

## Argument Reference

This resource supports the following arguments:

* `cidr_block` - (Optional) The IPv4 CIDR block for the VPC. CIDR can be explicitly set or it can be derived from IPAM using `ipv4_netmask_length`.
* `ipv4_ipam_pool_id` - (Optional) The ID of an IPv4 IPAM pool you want to use for allocating this VPC's CIDR. IPAM is a VPC feature that you can use to automate your IP address management workflows including assigning, tracking, troubleshooting, and auditing IP addresses across AWS Regions and accounts. Using IPAM you can monitor IP address usage throughout your AWS Organization.
* `ipv4_netmask_length` - (Optional) The netmask length of the IPv4 CIDR you want to allocate to this VPC. Requires specifying a `ipv4_ipam_pool_id`.
* `vpc_id` - (Required) The ID of the VPC to make the association with.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC CIDR association

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_vpc_ipv4_cidr_block_association` using the VPC CIDR Association ID. For example:

```terraform
import {
  to = aws_vpc_ipv4_cidr_block_association.example
  id = "vpc-cidr-assoc-xxxxxxxx"
}
```

Using `terraform import`, import `aws_vpc_ipv4_cidr_block_association` using the VPC CIDR Association ID. For example:

```console
% terraform import aws_vpc_ipv4_cidr_block_association.example vpc-cidr-assoc-xxxxxxxx
```

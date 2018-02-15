---
layout: "aws"
page_title: "AWS: aws_vpc_associate_cidr_block"
sidebar_current: "docs-aws-resource-vpc-associate-cidr-block"
description: |-
  Associates a CIDR block to a VPC
---

# aws_vpc_associate_cidr_block

Associates a CIDR block to a VPC

## Example Usage

IPv4 CIDR Block Association:

```hcl
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_associate_cidr_block" "secondary_cidr" {
  vpc_id = "${aws_vpc.main.id}"
  ipv4_cidr_block = "172.2.0.0/16"
}
```

IPv6 CIDR Block Association:

```hcl
resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_associate_cidr_block" "secondary_ipv6_cidr" {
  vpc_id = "${aws_vpc.main.id}"
  assign_generated_ipv6_cidr_block = true
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of the VPC to make the association with
* `cidr_block` - (Optional) The IPv4 CIDR block to associate to the VPC.
* `assign_generated_ipv6_cidr_block` - (Optional) Requests an Amazon-provided IPv6 CIDR 
block with a /56 prefix length for the VPC. You cannot specify the range of IP addresses, or 
the size of the CIDR block. Default is `false`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the VPC CIDR Association
* `ipv6_cidr_block` - The IPv6 CIDR block.


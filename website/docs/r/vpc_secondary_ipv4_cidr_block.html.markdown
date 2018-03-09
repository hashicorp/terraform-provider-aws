---
layout: "aws"
page_title: "AWS: aws_vpc_secondary_ipv4_cidr_block"
sidebar_current: "docs-aws-resource-vpc-secondary-ipv4-cidr-block"
description: |-
  Associate a secondary IPv4 CIDR blocks with a VPC
---

# aws_vpc_secondary_ipv4_cidr_block

Provides a resource to associate a secondary IPv4 CIDR blocks with a VPC.

## Example Usage

```hcl
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_secondary_ipv4_cidr_block" "secondary_cidr" {
  vpc_id = "${aws_vpc.main.id}"
  ipv4_cidr_block = "172.2.0.0/16"
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of the VPC to make the association with.
* `ipv4_cidr_block` - (Required) The secondary IPv4 CIDR block to associate with the VPC.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the VPC CIDR association

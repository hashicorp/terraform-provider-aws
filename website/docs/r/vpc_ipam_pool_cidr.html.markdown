---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool_cidr"
description: |-
  Provisions a CIDR from an IPAM address pool.
---

# Resource: aws_vpc_ipam_pool_cidr

Provisions a CIDR from an IPAM address pool.

## Example Usage

Basic usage:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
	operating_regions {
		region_name = data.aws_region.current.name
	}
}

resource "aws_vpc_ipam_pool" "test" {
	address_family = "ipv4"
	ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
	locale         = data.aws_region.current.name
}

resource "aws_vpc_ipam_pool_cidr" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "172.2.0.0/16"
}
```

Provision Public IPv6 Pool CIDRs:

```terraform
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
	operating_regions {
		region_name = data.aws_region.current.name
	}
}

resource "aws_vpc_ipam_pool" "ipv6_test_public" {
  address_family = "ipv6"
  ipam_scope_id  = aws_vpc_ipam.test.public_default_scope_id
  locale         = "us-east-1"
  description    = "public ipv6"
  advertisable   = false
}

resource "aws_vpc_ipam_pool_cidr" "ipv6_test_public" {
  ipam_pool_id = aws_vpc_ipam_pool.ipv6_test_public.id
  cidr = var.ipv6_cidr
  cidr_authorization_context {
    message = var.message
    signature = var.signature
  }
}
```

## Argument Reference

The following arguments are supported:

* `cidr` - (Optional) The CIDR you want to assign to the pool.
* `cidr_authorization_context` - (Optional) A signed document that proves that you are authorized to bring the specified IP address range to Amazon using BYOIP. This is not stored in the state file. See [cidr_authorization_context](#cidr_authorization_context) for more information.
* `ipam_pool_id` - (Required) The ID of the pool to which you want to assign a CIDR.

### cidr_authorization_context

* `message` - (Optional) The plain-text authorization message for the prefix and account.
* `signature` - (Optional) The signed authorization message for the prefix and account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the IPAM Pool Cidr concatenated with the IPAM Pool ID.
* `state` - The state of the CIDR.

## Import

IPAMs can be imported using the `cidr_ipam-pool-id`, e.g.

```
$ terraform import aws_vpc_ipam_pool_cidr.test 172.2.0.0/24_ipam-pool-0e634f5a1517cccdc
```

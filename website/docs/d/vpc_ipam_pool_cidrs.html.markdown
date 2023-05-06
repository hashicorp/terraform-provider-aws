---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool_cidrs"
description: |-
    Returns cidrs provisioned into an IPAM pool.
---

# Data Source: aws_vpc_ipam_pool_cidrs

`aws_vpc_ipam_pool_cidrs` provides details about an IPAM pool.

This resource can prove useful when an ipam pool was shared to your account and you want to know all (or a filtered list) of the CIDRs that are provisioned into the pool.

## Example Usage

Basic usage:

```terraform
data "aws_vpc_ipam_pool_cidrs" "c" {
  ipam_pool_id = data.aws_vpc_ipam_pool.p.id
}

data "aws_vpc_ipam_pool" "p" {
  filter {
    name   = "description"
    values = ["*mypool*"]
  }

  filter {
    name   = "address-family"
    values = ["ipv4"]
  }
}
```

Filtering:

```terraform
data "aws_vpc_ipam_pool_cidrs" "c" {
  ipam_pool_id = "ipam-pool-123"
  filter {
    name   = "cidr"
    values = ["10.*"]
  }
}

locals {
  mycidrs = [for cidr in data.aws_vpc_ipam_pool_cidrs.c.ipam_pool_cidrs :
    cidr.cidr if
  cidr.state == "provisioned"]
}

resource "aws_ec2_managed_prefix_list" "pls" {
  name           = "IPAM Pool (${aws_vpc_ipam_pool.test.id}) Cidrs"
  address_family = "IPv4"
  max_entries    = length(local.mycidrs)

  dynamic "entry" {
    for_each = local.mycidrs
    content {
      cidr        = entry.value
      description = entry.value
    }
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
VPCs in the current region. The given filters must match exactly one
VPC whose data will be exported as attributes.

* `ipam_pool_id` - ID of the IPAM pool you would like the list of provisioned CIDRs.
* `filter` - Custom filter block as described below.

## Attributes Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected IPAM Pool CIDRs.

The following attribute is additionally exported:

* `ipam_pool_cidrs` - The CIDRs provisioned into the IPAM pool, described below.

### ipam_pool_cidrs

* `cidr` - A network CIDR.
* `state` - The provisioning state of that CIDR.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `1m`)

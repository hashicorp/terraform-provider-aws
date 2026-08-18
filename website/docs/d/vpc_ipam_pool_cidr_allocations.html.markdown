---
subcategory: "VPC IPAM (IP Address Manager)"
layout: "aws"
page_title: "AWS: aws_vpc_ipam_pool_cidr_allocations"
description: |-
  Returns cidr allocations provisioned into an IPAM pool.
---
# Data Source: aws_vpc_ipam_pool_cidr_allocations

`aws_vpc_ipam_pool_cidr_allocations` provides details about allocations in an IPAM pool.

This resource can prove useful when an ipam pool was shared to your account and you want to know all (or a filtered list) of the Allocations that are provisioned into the pool.

## Example Usage

### Basic Usage

```terraform
data "aws_vpc_ipam_pool_cidr_allocations" "c" {
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
data "aws_vpc_ipam_pool_cidr_allocations" "c" {
  ipam_pool_id = "ipam-pool-123"
  filter {
    name   = "cidr"
    values = ["10.*"]
  }
}

locals {
  myallocations = {for allocation in data.aws_vpc_ipam_pool_cidr_allocations.c.ipam_pool_allocations :
    allocation.cidr => allocation.description
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
VPCs in the current region. The given filters must match exactly one
VPC whose data will be exported as attributes.

* `ipam_pool_id` - ID of the IPAM pool you would like the list of provisioned CIDRs.
* `filter` - Custom filter block as described below.

## Attribute Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected IPAM Pool Allocations.

The following attribute is additionally exported:

* `ipam_pool_allocations` - The CIDRs provisioned into the IPAM pool, described below.

### ipam_pool_cidrs

* `cidr` -  A network CIDR.
* `ipam_pool_allocation_id` - The ID of an allocation.
* `description` - A description of the pool allocation.
* `resource_id` - The ID of the resource.
* `resource_type` - The type of the resource.
* `resource_region` - The Amazon Web Services Region of the resource.
* `resource_owner` - The owner of the resource.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `1m`)
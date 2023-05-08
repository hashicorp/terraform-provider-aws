---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc"
description: |-
    Provides details about a specific VPC
---

# Data Source: aws_vpc

`aws_vpc` provides details about a specific VPC.

This resource can prove useful when a module accepts a vpc id as
an input variable and needs to, for example, determine the CIDR block of that
VPC.

## Example Usage

The following example shows how one might accept a VPC id as a variable
and use this data source to obtain the data necessary to create a subnet
within it.

```terraform
variable "vpc_id" {}

data "aws_vpc" "selected" {
  id = var.vpc_id
}

resource "aws_subnet" "example" {
  vpc_id            = data.aws_vpc.selected.id
  availability_zone = "us-west-2a"
  cidr_block        = cidrsubnet(data.aws_vpc.selected.cidr_block, 4, 1)
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
VPCs in the current region. The given filters must match exactly one
VPC whose data will be exported as attributes.

* `cidr_block` - (Optional) Cidr block of the desired VPC.

* `dhcp_options_id` - (Optional) DHCP options id of the desired VPC.

* `default` - (Optional) Boolean constraint on whether the desired VPC is
  the default VPC for the region.

* `filter` - (Optional) Custom filter block as described below.

* `id` - (Optional) ID of the specific VPC to retrieve.

* `state` - (Optional) Current state of the desired VPC.
  Can be either `"pending"` or `"available"`.

* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired VPC.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcs.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected VPC.

The following attribute is additionally exported:

* `arn` - ARN of VPC
* `enable_dns_support` - Whether or not the VPC has DNS support
* `enable_network_address_usage_metrics` - Whether Network Address Usage metrics are enabled for your VPC
* `enable_dns_hostnames` - Whether or not the VPC has DNS hostname support
* `instance_tenancy` - Allowed tenancy of instances launched into the
  selected VPC. May be any of `"default"`, `"dedicated"`, or `"host"`.
* `ipv6_association_id` - Association ID for the IPv6 CIDR block.
* `ipv6_cidr_block` - IPv6 CIDR block.
* `main_route_table_id` - ID of the main route table associated with this VPC.
* `owner_id` - ID of the AWS account that owns the VPC.

`cidr_block_associations` is also exported with the following attributes:

* `association_id` - Association ID for the IPv4 CIDR block.
* `cidr_block` - CIDR block for the association.
* `state` - State of the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

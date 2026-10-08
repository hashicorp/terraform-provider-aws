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

This data source supports the following arguments:

* `cidr_block` - (Optional) CIDR block of the desired VPC.
* `default` - (Optional) Boolean constraint on whether the desired VPC is the default VPC for the region.
* `dhcp_options_id` - (Optional) DHCP options id of the desired VPC.
* `filter` - (Optional) Custom filter block as described below. See [`filter` Block](#filter-block) below.
* `id` - (Optional) ID of the specific VPC to retrieve.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `state` - (Optional) Current state of the desired VPC. Can be either `"pending"` or `"available"`.
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired VPC.

### `filter` Block

The `filter` block supports the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcs.html).
* `values` - (Required) Set of values that are accepted for the given field. A VPC will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of VPC.
* `cidr_block_associations` - Information about the IPv4 CIDR blocks associated with the VPC. See [`cidr_block_associations` Block](#cidr_block_associations-block) below.
* `enable_dns_hostnames` - Whether the VPC has DNS hostname support.
* `enable_dns_support` - Whether the VPC has DNS support.
* `enable_network_address_usage_metrics` - Whether Network Address Usage metrics are enabled for your VPC.
* `instance_tenancy` - Allowed tenancy of instances launched into the selected VPC. May be any of `"default"`, `"dedicated"`, or `"host"`.
* `ipv6_association_id` - (**Deprecated** use `ipv6_cidr_block_associations` instead) Association ID for the IPv6 CIDR block.
* `ipv6_cidr_block` - (**Deprecated** use `ipv6_cidr_block_associations` instead) IPv6 CIDR block.
* `ipv6_cidr_block_associations` - Information about the IPv6 CIDR blocks associated with the VPC. See [`ipv6_cidr_block_associations` Block](#ipv6_cidr_block_associations-block) below.
* `main_route_table_id` - ID of the main route table associated with this VPC.
* `owner_id` - ID of the AWS account that owns the VPC.

### `cidr_block_associations` Block

The `cidr_block_associations` block exports the following attributes:

* `association_id` - Association ID for the IPv4 CIDR block.
* `cidr_block` - CIDR block for the association.
* `state` - State of the association.

### `ipv6_cidr_block_associations` Block

The `ipv6_cidr_block_associations` block exports the following attributes:

* `association_id` - Association ID for the IPv4 CIDR block.
* `ip_source` - Source that allocated the IP address space. Values: `amazon`, `byoip`, `none`.
* `ipv6_address_attribute` - Whether the address is `public` or `private`.
* `ipv6_cidr_block` - IPv6 CIDR block for the association.
* `ipv6_pool` - Name of IPv6 address pool from which the IPv6 CIDR block is allocated.
* `network_border_group` - Name of association's network border group.
* `state` - State of the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_nat_gateway"
description: |-
    Provides details about a specific Nat Gateway
---

# Data Source: aws_nat_gateway

Provides details about a specific Nat Gateway.

## Example Usage

```terraform
data "aws_nat_gateway" "default" {
  subnet_id = aws_subnet.public.id
}
```

Usage with tags:

```terraform
data "aws_nat_gateway" "default" {
  subnet_id = aws_subnet.public.id

  tags = {
    Name = "gw NAT"
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Nat Gateways in the current region. The given filters must match exactly one
Nat Gateway whose data will be exported as attributes.

* `id` - (Optional) ID of the specific Nat Gateway to retrieve.
* `subnet_id` - (Optional) ID of subnet that the Nat Gateway resides in.
* `vpc_id` - (Optional) ID of the VPC that the Nat Gateway resides in.
* `state` - (Optional) State of the NAT gateway (pending | failed | available | deleting | deleted ).
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired Nat Gateway.
* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNatGateways.html).
* `values` - (Required) Set of values that are accepted for the given field.
  An Nat Gateway will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` block are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected Nat Gateway.

`addresses` are also exported with the following attributes, when they are relevant:
Each attachment supports the following:

* `allocation_id` - ID of the EIP allocated to the selected Nat Gateway.
* `connectivity_type` - Connectivity type of the NAT Gateway.
* `network_interface_id` - The ID of the ENI allocated to the selected Nat Gateway.
* `private_ip` - Private Ip address of the selected Nat Gateway.
* `public_ip` - Public Ip (EIP) address of the selected Nat Gateway.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

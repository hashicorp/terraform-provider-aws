---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_nat_gateway"
description: |-
    Provides details about a specific VPC NAT Gateway.
---

# Data Source: aws_nat_gateway

Provides details about a specific VPC NAT Gateway.

## Example Usage

```terraform
data "aws_nat_gateway" "default" {
  subnet_id = aws_subnet.public.id
}
```

### With tags

```terraform
data "aws_nat_gateway" "default" {
  subnet_id = aws_subnet.public.id

  tags = {
    Name = "gw NAT"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `id` - (Optional) ID of the specific NAT Gateway to retrieve.
* `subnet_id` - (Optional) ID of subnet that the NAT Gateway resides in.
* `vpc_id` - (Optional) ID of the VPC that the NAT Gateway resides in.
* `state` - (Optional) State of the NAT Gateway (pending | failed | available | deleting | deleted ).
* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired NAT Gateway.
* `filter` - (Optional) Custom filter block as described below.

The arguments of this data source act as filters for querying the available
NAT Gateways in the current Region. The given filters must match exactly one
NAT Gateway whose data will be exported as attributes.

### `filter`

More complex filters can be expressed using one or more `filter` sub-blocks, which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNatGateways.html).
* `values` - (Required) Set of values that are accepted for the given field.
  An Nat Gateway will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `allocation_id` - ID of the EIP allocated to the selected NAT Gateway.
* `association_id` - The association ID of the Elastic IP address that's associated with the NAT Gateway. Only available when `connectivity_type` is `public`.
* `connectivity_type` - Connectivity type of the NAT Gateway.
* `network_interface_id` - The ID of the ENI allocated to the selected NAT Gateway.
* `private_ip` - Private IP address of the selected NAT Gateway.
* `public_ip` - Public IP (EIP) address of the selected NAT Gateway.
* `secondary_allocation_ids` - Secondary allocation EIP IDs for the selected NAT Gateway.
* `secondary_private_ip_address_count` - The number of secondary private IPv4 addresses assigned to the selected NAT Gateway.
* `secondary_private_ip_addresses` - Secondary private IPv4 addresses assigned to the selected NAT Gateway.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

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

* `allocation_id` - (zonal NAT gateways only) ID of the EIP allocated to the selected NAT Gateway.
* `association_id` - (zonal NAT gateways only) The association ID of the Elastic IP address that's associated with the NAT Gateway. Only available when `connectivity_type` is `public`.
* `auto_provision_zones` - (regional NAT gateways only) Indicates whether AWS automatically manages AZ coverage.
* `auto_scaling_ips` - (regional NAT gateways only) Indicates whether AWS automatically allocates additional Elastic IP addresses (EIPs) in an AZ when the NAT gateway needs more ports due to increased concurrent connections to a single destination from that AZ.
* `availability_mode` - Specifies whether to create a zonal (single-AZ) or regional (multi-AZ) NAT gateway.
* `availability_zone_address` - (regional NAT gateways only) Repeatable configuration block for the Elastic IP addresses (EIPs) and availability zones for the regional NAT gateway.
    * `allocation_ids` - List of allocation IDs of the Elastic IP addresses (EIPs) to be used for handling outbound NAT traffic in this specific Availability Zone.
    * `availability_zone` - Availability Zone (e.g. `us-west-2a`) where this specific NAT gateway configuration will be active.
    * `availability_zone_id` - Availability Zone ID (e.g. `usw2-az2`) where this specific NAT gateway configuration will be active
* `connectivity_type` - Connectivity type of the NAT Gateway.
* `network_interface_id` - (zonal NAT gateways only) The ID of the ENI allocated to the selected NAT Gateway.
* `private_ip` - (zonal NAT gateways only) Private IP address of the selected NAT Gateway.
* `public_ip` - (zonal NAT gateways only) Public IP (EIP) address of the selected NAT Gateway.
* `regional_nat_gateway_address` - (regional NAT gateways only) Repeatable blocks for information about the IP addresses and network interface associated with the regional NAT gateway.
    * `allocation_id` - Allocation ID of the Elastic IP address.
    * `association_id` - Association ID of the Elastic IP address.
    * `availability_zone` - Availability Zone where this specific NAT gateway configuration is active.
    * `availability_zone_id` - Availability Zone ID where this specific NAT gateway configuration is active
    * `network_interface_id` - ID of the network interface.
    * `public_ip` - Public IP address.
    * `status` - Status of the NAT gateway address.
* `route_table_id` - (regional NAT gateways only) ID of the automatically created route table.
* `secondary_allocation_ids` - (zonal NAT gateways only) Secondary allocation EIP IDs for the selected NAT Gateway.
* `secondary_private_ip_address_count` - (zonal NAT gateways only) The number of secondary private IPv4 addresses assigned to the selected NAT Gateway.
* `secondary_private_ip_addresses` - (zonal NAT gateways only) Secondary private IPv4 addresses assigned to the selected NAT Gateway.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

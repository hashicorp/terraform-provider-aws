---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_route"
description: |-
    Provides details about a specific Route
---

# Data Source: aws_route

`aws_route` provides details about a specific Route.

This resource can prove useful when finding the resource associated with a CIDR. For example, finding the peering connection associated with a CIDR value.

## Example Usage

The following example shows how one might use a CIDR value to find a network interface id and use this to create a data source of that network interface.

```terraform
variable "subnet_id" {}

data "aws_route_table" "selected" {
  subnet_id = var.subnet_id
}

data "aws_route" "route" {
  route_table_id         = aws_route_table.selected.id
  destination_cidr_block = "10.0.1.0/24"
}

data "aws_network_interface" "interface" {
  id = data.aws_route.route.network_interface_id
}
```

## Argument Reference

This data source supports the following arguments:

* `route_table_id` - (Required) ID of the specific Route Table containing the Route entry.
* `carrier_gateway_id` - (Optional) EC2 Carrier Gateway ID of the Route belonging to the Route Table.
* `core_network_arn` - (Optional) Core network ARN of the Route belonging to the Route Table.
* `destination_cidr_block` - (Optional) CIDR block of the Route belonging to the Route Table.
* `destination_ipv6_cidr_block` - (Optional) IPv6 CIDR block of the Route belonging to the Route Table.
* `destination_prefix_list_id` - (Optional) ID of a [managed prefix list](ec2_managed_prefix_list.html) destination of the Route belonging to the Route Table.
* `egress_only_gateway_id` - (Optional) Egress Only Gateway ID of the Route belonging to the Route Table.
* `gateway_id` - (Optional) Gateway ID of the Route belonging to the Route Table.
* `instance_id` - (Optional) Instance ID of the Route belonging to the Route Table.
* `local_gateway_id` - (Optional) Local Gateway ID of the Route belonging to the Route Table.
* `nat_gateway_id` - (Optional) NAT Gateway ID of the Route belonging to the Route Table.
* `network_interface_id` - (Optional) Network Interface ID of the Route belonging to the Route Table.
* `transit_gateway_id` - (Optional) EC2 Transit Gateway ID of the Route belonging to the Route Table.
* `vpc_peering_connection_id` - (Optional) VPC Peering Connection ID of the Route belonging to the Route Table.

The arguments of this data source act as filters for querying the available Route in the current region. The given filters must match exactly oneRoute whose data will be exported as attributes.

## Attribute Reference

This data source exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

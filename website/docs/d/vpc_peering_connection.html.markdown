---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_peering_connection"
description: |-
    Provides details about a specific VPC peering connection.
---

# Data Source: aws_vpc_peering_connection

The VPC Peering Connection data source provides details about
a specific VPC peering connection.

## Example Usage

```terraform
# Declare the data source
data "aws_vpc_peering_connection" "pc" {
  vpc_id          = aws_vpc.foo.id
  peer_cidr_block = "10.0.1.0/22"
}

# Create a route table
resource "aws_route_table" "rt" {
  vpc_id = aws_vpc.foo.id
}

# Create a route
resource "aws_route" "r" {
  route_table_id            = aws_route_table.rt.id
  destination_cidr_block    = data.aws_vpc_peering_connection.pc.peer_cidr_block
  vpc_peering_connection_id = data.aws_vpc_peering_connection.pc.id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC peering connection.
The given filters must match exactly one VPC peering connection whose data will be exported as attributes.

* `id` - (Optional) ID of the specific VPC Peering Connection to retrieve.

* `status` - (Optional) Status of the specific VPC Peering Connection to retrieve.

* `vpc_id` - (Optional) ID of the requester VPC of the specific VPC Peering Connection to retrieve.

* `owner_id` - (Optional) AWS account ID of the owner of the requester VPC of the specific VPC Peering Connection to retrieve.

* `cidr_block` - (Optional) Primary CIDR block of the requester VPC of the specific VPC Peering Connection to retrieve.

* `region` - (Optional) Region of the requester VPC of the specific VPC Peering Connection to retrieve.

* `peer_vpc_id` - (Optional) ID of the accepter VPC of the specific VPC Peering Connection to retrieve.

* `peer_owner_id` - (Optional) AWS account ID of the owner of the accepter VPC of the specific VPC Peering Connection to retrieve.

* `peer_cidr_block` - (Optional) Primary CIDR block of the accepter VPC of the specific VPC Peering Connection to retrieve.

* `peer_region` - (Optional) Region of the accepter VPC of the specific VPC Peering Connection to retrieve.

* `filter` - (Optional) Custom filter block as described below.

* `tags` - (Optional) Map of tags, each pair of which must exactly match
  a pair on the desired VPC Peering Connection.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) Name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcPeeringConnections.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC Peering Connection will be selected if any one of the given values matches.

## Attribute Reference

All of the argument attributes except `filter` are also exported as result attributes.

* `accepter` - Configuration block that describes [VPC Peering Connection]
(https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html) options set for the accepter VPC.

* `cidr_block_set` - List of objects with IPv4 CIDR blocks of the requester VPC.

* `ipv6_cidr_block_set` - List of objects with IPv6 CIDR blocks of the requester VPC.

* `peer_cidr_block_set` - List of objects with IPv4 CIDR blocks of the accepter VPC.

* `peer_ipv6_cidr_block_set` - List of objects with IPv6 CIDR blocks of the accepter VPC.

* `requester` - Configuration block that describes [VPC Peering Connection]
(https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html) options set for the requester VPC.

#### Accepter and Requester Attribute Reference

* `allow_remote_vpc_dns_resolution` - Indicates whether a local VPC can resolve public DNS hostnames to
private IP addresses when queried from instances in a peer VPC.

#### CIDR block set Attribute Reference

* `cidr_block` - CIDR block associated to the VPC of the specific VPC Peering Connection.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

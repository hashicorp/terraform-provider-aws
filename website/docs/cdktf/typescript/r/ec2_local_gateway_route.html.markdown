---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route"
description: |-
  Manages an EC2 Local Gateway Route
---

# Resource: aws_ec2_local_gateway_route

Manages an EC2 Local Gateway Route. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-networking-components.html#routing).

## Example Usage

```terraform
resource "aws_ec2_local_gateway_route" "example" {
  destination_cidr_block                   = "172.16.0.0/16"
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.example.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.example.id
}
```

## Argument Reference

The following arguments are required:

* `destinationCidrBlock` - (Required) IPv4 CIDR range used for destination matches. Routing decisions are based on the most specific match.
* `localGatewayRouteTableId` - (Required) Identifier of EC2 Local Gateway Route Table.
* `localGatewayVirtualInterfaceGroupId` - (Required) Identifier of EC2 Local Gateway Virtual Interface Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Local Gateway Route Table identifier and destination CIDR block separated by underscores (`_`)

## Import

`awsEc2LocalGatewayRoute` can be imported by using the EC2 Local Gateway Route Table identifier and destination CIDR block separated by underscores (`_`), e.g.,

```
$ terraform import aws_ec2_local_gateway_route.example lgw-rtb-12345678_172.16.0.0/16
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-1e91a283bb3d9d11b1b4699831f824ebe2b6e4d62ba4a95fea8fbf7e64228942 -->
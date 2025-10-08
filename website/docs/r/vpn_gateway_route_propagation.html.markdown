---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_gateway_route_propagation"
description: |-
  Requests automatic route propagation between a VPN gateway and a route table.
---

# Resource: aws_vpn_gateway_route_propagation

Requests automatic route propagation between a VPN gateway and a route table.

~> **Note:** This resource should not be used with a route table that has
the `propagating_vgws` argument set. If that argument is set, any route
propagation not explicitly listed in its value will be removed.

## Example Usage

```terraform
resource "aws_vpn_gateway_route_propagation" "example" {
  vpn_gateway_id = aws_vpn_gateway.example.id
  route_table_id = aws_route_table.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpn_gateway_id` - The id of the `aws_vpn_gateway` to propagate routes from.
* `route_table_id` - The id of the `aws_route_table` to propagate routes into.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `2m`)
- `delete` - (Default `2m`)

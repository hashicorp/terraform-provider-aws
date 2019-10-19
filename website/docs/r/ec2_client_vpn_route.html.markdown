---
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_route"
description: |-
  Provides route resource for AWS Client VPN endpoints.
---

# Resource: aws_ec2_client_vpn_route

Provides routes resource for AWS Client VPN endpoints. For more information on usage, please see the 
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```hcl
resource "aws_ec2_client_vpn_route" "example" {
  client_vpn_endpoint_id    = "${aws_ec2_client_vpn_endpoint.example.id}"
  client_vpn_association_id = "${aws_ec2_client_vpn_network_association.example.id}"
  cidr_block                = "10.202.0.0/16"
}
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Required) The CIDR block for the subnet.
* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `client_vpn_association_id` - (Required) The ID of the Client VPN network association.
* `description` - (Optional) The description of the route.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the route.
* `subnet_id` - The ID of the associated subnet.

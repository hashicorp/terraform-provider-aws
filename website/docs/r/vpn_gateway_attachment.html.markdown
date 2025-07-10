---
subcategory: "VPN (Site-to-Site)"
layout: "aws"
page_title: "AWS: aws_vpn_gateway_attachment"
description: |-
  Provides a Virtual Private Gateway attachment resource.
---

# Resource: aws_vpn_gateway_attachment

Provides a Virtual Private Gateway attachment resource, allowing for an existing
hardware VPN gateway to be attached and/or detached from a VPC.

-> **Note:** The [`aws_vpn_gateway`](vpn_gateway.html)
resource can also automatically attach the Virtual Private Gateway it creates
to an existing VPC by setting the [`vpc_id`](vpn_gateway.html#vpc_id) attribute accordingly.

## Example Usage

```terraform
resource "aws_vpc" "network" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpn_gateway" "vpn" {
  tags = {
    Name = "example-vpn-gateway"
  }
}

resource "aws_vpn_gateway_attachment" "vpn_attachment" {
  vpc_id         = aws_vpc.network.id
  vpn_gateway_id = aws_vpn_gateway.vpn.id
}
```

See [Virtual Private Cloud](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Introduction.html)
and [Virtual Private Gateway](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_VPN.html) user
guides for more information.

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_id` - (Required) The ID of the VPC.
* `vpn_gateway_id` - (Required) The ID of the Virtual Private Gateway.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `vpc_id` - The ID of the VPC that Virtual Private Gateway is attached to.
* `vpn_gateway_id` - The ID of the Virtual Private Gateway.

## Import

You cannot import this resource.

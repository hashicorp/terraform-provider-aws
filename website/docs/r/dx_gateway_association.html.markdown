---
layout: "aws"
page_title: "AWS: aws_dx_gateway_association"
sidebar_current: "docs-aws-resource-dx-gateway-association"
description: |-
  Associates a Direct Connect Gateway with a VGW.
---

# Resource: aws_dx_gateway_association

Associates a Direct Connect Gateway with a VGW. For creating cross-account association proposals, see the [`aws_dx_gateway_association_proposal` resource](/docs/providers/aws/r/dx_gateway_association_proposal.html).

## Example Usage

### Basic

```hcl
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_vpc" "example" {
  cidr_block = "10.255.255.0/28"
}

resource "aws_vpn_gateway" "example" {
  vpc_id = "${aws_vpc.example.id}"
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id  = "${aws_dx_gateway.example.id}"
  vpn_gateway_id = "${aws_vpn_gateway.example.id}"
}
```

### Allowed Prefixes

```hcl
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_vpc" "example" {
  cidr_block = "10.255.255.0/28"
}

resource "aws_vpn_gateway" "example" {
  vpc_id = "${aws_vpc.example.id}"
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id  = "${aws_dx_gateway.example.id}"
  vpn_gateway_id = "${aws_vpn_gateway.example.id}"

  allowed_prefixes = [
    "210.52.109.0/24",
    "175.45.176.0/22",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `dx_gateway_id` - (Required) The ID of the Direct Connect gateway.
* `vpn_gateway_id` - (Required) The ID of the VGW with which to associate the gateway.
* `allowed_prefixes` - (Optional) VPC prefixes (CIDRs) to advertise to the Direct Connect gateway. Defaults to the CIDR block of the VPC associated with the Virtual Gateway. To enable drift detection, must be configured.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Direct Connect gateway association resource.
* `dx_gateway_association_id` - The ID of the Direct Connect gateway association.

## Timeouts

`aws_dx_gateway_association` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `15 minutes`) Used for creating the association
- `update` - (Default `10 minutes`) Used for updating the association
- `delete` - (Default `15 minutes`) Used for destroying the association

## Import

Direct Connect gateway associations can be imported using `dx_gateway_id` together with `vpn_gateway_id`,
e.g.

```
$ terraform import aws_dx_gateway_association.example dxgw-12345678/vgw-98765432
```

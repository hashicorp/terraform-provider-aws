---
layout: "aws"
page_title: "AWS: aws_dx_gateway_association_proposal"
sidebar_current: "docs-aws-resource-dx-gateway-association-proposal"
description: |-
  Manages a Direct Connect Gateway Association Proposal.
---

# Resource: aws_dx_gateway_association_proposal

Manages a Direct Connect Gateway Association Proposal, typically for enabling cross-account associations. For single account associations, see the [`aws_dx_gateway_association` resource](/docs/providers/aws/r/dx_gateway_association.html).

## Example Usage

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

resource "aws_dx_gateway_association_proposal" "example" {
  dx_gateway_id               = "${aws_dx_gateway.example.id}"
  dx_gateway_owner_account_id = "${aws_dx_gateway.example.owner_account_id}"
  vpn_gateway_id              = "${aws_vpn_gateway.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `dx_gateway_id` - (Required) Direct Connect Gateway identifier.
* `dx_gateway_owner_account_id` - (Required) AWS Account identifier of the Direct Connect Gateway.
* `vpn_gateway_id` - (Required) Virtual Gateway identifier to associate with the Direct Connect Gateway.
* `allowed_prefixes` - (Optional) VPC prefixes (CIDRs) to advertise to the Direct Connect gateway. Defaults to the CIDR block of the VPC associated with the Virtual Gateway. To enable drift detection, must be configured.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Direct Connect Gateway Association Proposal identifier

## Import

Direct Connect Gateway Association Proposals can be imported using the proposal ID, e.g.

```
$ terraform import aws_dx_gateway_association_proposal.example ac90e981-b718-4364-872d-65478c84fafe
```

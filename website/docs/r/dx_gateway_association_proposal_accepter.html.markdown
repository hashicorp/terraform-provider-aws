---
layout: "aws"
page_title: "AWS: aws_dx_gateway_association_proposal_accepter"
sidebar_current: "docs-aws-resource-dx-gateway-association-proposal-accepter"
description: |-
  Provides a resource to manage the accepter's side of a Direct Connect Gateway Association Proposal.
---

# Resource: aws_dx_gateway_association_proposal_accepter

Provides a resource to manage the accepter's side of a Direct Connect Gateway Association Proposal.
This resource accepts a proposal request (created by another AWS account) to attach a virtual private gateway
(owned by the proposal creator's account) to a Direct Connect gateway (owned by the proposal accepter's account).

## Example Usage

```hcl
provider "aws" {
  # Creator's credentials.
}

provider "aws" {
  alias = "accepter"

  # Accepter's credentials.
}

# Creator's side of the proposal.
data "aws_caller_identity" "creator" {}

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

# Accepter's side of the proposal.
resource "aws_dx_gateway" "example" {
  provider = "aws.accepter"

  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_dx_gateway_association_proposal_accepter" "example" {
  provider = "aws.accepter"

  proposal_id                  = "${aws_dx_gateway_association_proposal.example.id}"
  dx_gateway_id                = "${aws_dx_gateway.example.id}"
  vpn_gateway_owner_account_id = "${data.aws_caller_identity.creator.account_id}"
}
```

## Argument Reference

The following arguments are supported:

* `dx_gateway_id` - (Required) Direct Connect Gateway identifier.
* `proposal_id` - (Required) Direct Connect Gateway Association Proposal identifier.
* `vpn_gateway_owner_account_id` - (Required) AWS Account identifier of the Virtual Gateway associated with the Direct Connect Gateway.
* `override_allowed_prefixes` - (Optional) Prefixes (CIDRs) that override the VPC prefixes (CIDRs) advertised to the Direct Connect gateway.

### Removing `aws_dx_gateway_association_proposal_accepter` from your configuration

The owner of the virtual private gateway can delete the Direct Connect gateway association proposal if it is still pending acceptance.
After an association proposal is accepted, you can't delete it, but you can disassociate the virtual private gateway from the Direct Connect gateway.
Removing a `aws_dx_gateway_association_proposal_accepter` resource from your configuration will remove it
from your statefile and management, **but will not delete the Direct Connect gateway association proposal.**

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Direct Connect Gateway Association Proposal identifier

## Import

Direct Connect Gateway Association Proposal Accepters can be imported using the proposal ID, e.g.

```
$ terraform import aws_dx_gateway_association_proposal_accepter.example ac90e981-b718-4364-872d-65478c84fafe
```

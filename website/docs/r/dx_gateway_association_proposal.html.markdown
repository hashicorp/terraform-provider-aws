---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_gateway_association_proposal"
description: |-
  Manages a Direct Connect Gateway Association Proposal.
---

# Resource: aws_dx_gateway_association_proposal

Manages a Direct Connect Gateway Association Proposal, typically for enabling cross-account associations. For single account associations, see the [`aws_dx_gateway_association` resource](/docs/providers/aws/r/dx_gateway_association.html).

## Example Usage

```terraform
resource "aws_dx_gateway_association_proposal" "example" {
  dx_gateway_id               = aws_dx_gateway.example.id
  dx_gateway_owner_account_id = aws_dx_gateway.example.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.example.id
}
```

A full example of how to create a VPN Gateway in one AWS account, create a Direct Connect Gateway in a second AWS account, and associate the VPN Gateway with the Direct Connect Gateway via the `aws_dx_gateway_association_proposal` and `aws_dx_gateway_association` resources can be found in [the `./examples/dx-gateway-cross-account-vgw-association` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/dx-gateway-cross-account-vgw-association).

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `associated_gateway_id` - (Required) The ID of the VGW or transit gateway with which to associate the Direct Connect gateway.
* `dx_gateway_id` - (Required) Direct Connect Gateway identifier.
* `dx_gateway_owner_account_id` - (Required) AWS Account identifier of the Direct Connect Gateway's owner.
* `allowed_prefixes` - (Optional) VPC prefixes (CIDRs) to advertise to the Direct Connect gateway. Defaults to the CIDR block of the VPC associated with the Virtual Gateway. To enable drift detection, must be configured.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Direct Connect Gateway Association Proposal identifier.
* `associated_gateway_owner_account_id` - The ID of the AWS account that owns the VGW or transit gateway with which to associate the Direct Connect gateway.
* `associated_gateway_type` - The type of the associated gateway, `transitGateway` or `virtualPrivateGateway`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Direct Connect Gateway Association Proposals using either a proposal ID or proposal ID, Direct Connect Gateway ID and associated gateway ID separated by `/`. For example:

Using a proposal ID:

```terraform
import {
  to = aws_dx_gateway_association_proposal.example
  id = "ac90e981-b718-4364-872d-65478c84fafe"
}
```

Using a proposal ID, Direct Connect Gateway ID and associated gateway ID separated by `/`:

```terraform
import {
  to = aws_dx_gateway_association_proposal.example
  id = "ac90e981-b718-4364-872d-65478c84fafe/abcd1234-dcba-5678-be23-cdef9876ab45/vgw-12345678"
}
```

**With `terraform import`**, import Direct Connect Gateway Association Proposals using either a proposal ID or proposal ID, Direct Connect Gateway ID and associated gateway ID separated by `/`. For example:

Using a proposal ID:

```console
% terraform import aws_dx_gateway_association_proposal.example ac90e981-b718-4364-872d-65478c84fafe
```

Using a proposal ID, Direct Connect Gateway ID and associated gateway ID separated by `/`:

```console
% terraform import aws_dx_gateway_association_proposal.example ac90e981-b718-4364-872d-65478c84fafe/abcd1234-dcba-5678-be23-cdef9876ab45/vgw-12345678
```

The latter case is useful when a previous proposal has been accepted and deleted by AWS.
The `aws_dx_gateway_association_proposal` resource will then represent a pseudo-proposal for the same Direct Connect Gateway and associated gateway. If no previous proposal is available, use a tool like [`uuidgen`](http://manpages.ubuntu.com/manpages/bionic/man1/uuidgen.1.html) to generate a new random pseudo-proposal ID.

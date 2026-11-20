---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_peering_attachment"
description: |-
  Manages an EC2 Transit Gateway Peering Attachment
---

# Resource: aws_ec2_transit_gateway_peering_attachment

Manages an EC2 Transit Gateway Peering Attachment.
For examples of custom route table association and propagation, see the [EC2 Transit Gateway Networking Examples Guide](https://docs.aws.amazon.com/vpc/latest/tgw/TGW_Scenarios.html).

## Example Usage

```terraform
provider "aws" {
  alias  = "local"
  region = "us-east-1"
}

provider "aws" {
  alias  = "peer"
  region = "us-west-2"
}

data "aws_region" "peer" {
  provider = aws.peer
}

resource "aws_ec2_transit_gateway" "local" {
  provider = aws.local

  tags = {
    Name = "Local TGW"
  }
}

resource "aws_ec2_transit_gateway" "peer" {
  provider = aws.peer

  tags = {
    Name = "Peer TGW"
  }
}

resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  peer_account_id         = aws_ec2_transit_gateway.peer.owner_id
  peer_region             = data.aws_region.peer.region
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.local.id

  tags = {
    Name = "TGW Peering Requestor"
  }
}
```

Cross-Region usage with multipath resilience:

```terraform
resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  provider = aws.local

  peer_region             = data.aws_region.peer.region
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.local.id

  options {
    allow_multipath_resilience = true
  }
}

# Accepting the attachment also allows multipath resilience for the accepter.
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "example" {
  provider = aws.peer

  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.example.id
}
```

A full example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a to a Transit Gateway in the second account via the `aws_ec2_transit_gateway_peering_attachment` resource can be found in [the `./examples/transit-gateway-cross-account-peering-attachment` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/transit-gateway-cross-account-peering-attachment).

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `peer_account_id` - (Optional) Account ID of EC2 Transit Gateway to peer with. Defaults to the account ID the [AWS provider](/docs/providers/aws/index.html) is currently connected to.
* `peer_region` - (Required) Region of EC2 Transit Gateway to peer with.
* `peer_transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway to peer with.
* `options` - (Optional) Block describing the dynamic routing and multipath resilience options for the transit gateway peering request. If not configured, the value is populated from the API. See [options](#options) below for more details!
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Peering Attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.

### options

The `options` block supports the following:

* `allow_multipath_resilience` - (Optional) Whether to allow [multipath resilience](https://docs.aws.amazon.com/vpc/latest/tgw/enable-multipath-resilience-tgw.html) for the requester's side of a cross-Region attachment. Multipath resilience takes effect only when both sides allow it, and additional charges apply. If not set, the current setting is left unchanged. Can be changed only once the attachment is available.
* `dynamic_routing` - (Optional) Indicates whether dynamic routing is enabled or disabled. Supports `enable` and `disable`. Defaults to `disable`.
* `multipath_resilience_mode` - (Optional) Multipath resilience mode. Valid values: `always-on`. Can only be set when `allow_multipath_resilience` is `true`. Defaults to `always-on` when multipath resilience is allowed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `id` - EC2 Transit Gateway Attachment identifier.
* `multipath_resilience_status` - Multipath resilience status of the attachment. Values are `enabled` (both sides allow multipath resilience), `pending` (only one side allows it, or it is being enabled) and `disabled` (neither side allows it).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_peering_attachment` using the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_peering_attachment.example
  id = "tgw-attach-12345678"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_peering_attachment` using the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_peering_attachment.example tgw-attach-12345678
```

---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall_transit_gateway_attachment_accepter"
description: |-
  Manages an AWS Network Firewall Firewall Transit Gateway Attachment Accepter.
---

# Resource: aws_networkfirewall_firewall_transit_gateway_attachment_accepter

Manages an AWS Network Firewall Firewall Transit Gateway Attachment Accepter.

When a cross-account (requester's AWS account differs from the accepter's AWS account) requester creates a Network Firewall with Transit Gateway ID using `aws_networkfirewall_firewall`. Then an EC2 Transit Gateway VPC Attachment resource is automatically created in the accepter's account.
The accepter can use the `aws_networkfirewall_firewall_transit_gateway_attachment_accepter` resource to "adopt" its side of the connection into management.

~> **NOTE:** If the `transit_gateway_id` argument in the `aws_networkfirewall_firewall` resource is used to attach a firewall to a transit gateway in a cross-account setup (where **Auto accept shared attachments** is disabled), the resource will be considered created when the transit gateway attachment is in the *Pending Acceptance* state and the firewall is in the *Provisioning* status. At this point, you can use the `aws_networkfirewall_firewall_transit_gateway_attachment_accepter` resource to finalize the network firewall deployment. Once the transit gateway attachment reaches the *Available* state, the firewall status *Ready*.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkfirewall_firewall_transit_gateway_attachment_accepter" "example" {
  transit_gateway_attachment_id = aws_networkfirewall_firewall.example.firewall_status[0].transit_gateway_attachment_sync_states[0].attachment_id
}
```

For an example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, create a Network Firewall attached to the Transit Gateway in the second account, and accept the attachment, see the [`aws_networkfirewall_firewall` resource](/docs/providers/aws/r/networkfirewall_firewall.html#transit-gateway-attached-firewall-cross-account).

## Argument Reference

This resource supports the following arguments:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_attachment_id` - (Required) Unique identifier of the transit gateway attachment to accept. This ID is returned in the response when creating a transit gateway-attached firewall.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Firewall Firewall Transit Gateway Attachment Accepter using the `transit_gateway_attachment_id`. For example:

```terraform
import {
  to = aws_networkfirewall_firewall_transit_gateway_attachment_accepter.example
  id = "tgw-attach-0c3b7e9570eee089c"
}
```

Using `terraform import`, import Network Firewall Firewall Transit Gateway Attachment Accepter using the `transit_gateway_attachment_id`. For example:

```console
% terraform import aws_networkfirewall_firewall_transit_gateway_attachment_accepter.example tgw-attach-0c3b7e9570eee089c
```

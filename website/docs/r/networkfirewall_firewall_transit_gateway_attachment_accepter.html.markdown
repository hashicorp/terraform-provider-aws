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
  transit_gateway_attachment_id = aws_networkfirewall_firewall.example.firewall_status[0].transit_gateway_attachment_sync_state[0].attachment_id
}
```

A full example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, and create Network Firewall in the second account to the Transit Gateway via the `aws_networkfirewall_firewall` and `aws_networkfirewall_firewall_transit_gateway_attachment_accepter` resources can be found in [the `./examples/network-firewall-cross-account-transit-gateway` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/network-firewall-cross-account-transit-gateway)

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_attachment_id` - (Required) The unique identifier of the transit gateway attachment to accept. This ID is returned in the response when creating a transit gateway-attached firewall.

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

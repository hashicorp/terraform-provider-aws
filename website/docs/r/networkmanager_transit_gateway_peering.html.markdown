---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_peering"
description: |-
  Manages a Network Manager transit gateway peering connection.
---

# Resource: aws_networkmanager_transit_gateway_peering

Manages a Network Manager transit gateway peering connection. Creates a peering connection between an AWS Cloud WAN core network and an AWS Transit Gateway.

## Example Usage

```terraform
resource "aws_networkmanager_transit_gateway_peering" "example" {
  core_network_id     = awscc_networkmanager_core_network.example.id
  transit_gateway_arn = aws_ec2_transit_gateway.example.arn
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of a core network.
* `transit_gateway_arn` - (Required) ARN of the transit gateway for the peering request.

The following arguments are optional:

* `tags` - (Optional) Key-value tags for the peering. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Peering ARN.
* `core_network_arn` - ARN of the core network.
* `edge_location` - Edge location for the peer.
* `id` - Peering ID.
* `owner_account_id` - ID of the account owner.
* `peering_type` - Type of peering. This will be `TRANSIT_GATEWAY`.
* `resource_arn` - Resource ARN of the peer.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `transit_gateway_peering_attachment_id` - ID of the transit gateway peering attachment.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_transit_gateway_peering` using the peering ID. For example:

```terraform
import {
  to = aws_networkmanager_transit_gateway_peering.example
  id = "peering-444555aaabbb11223"
}
```

Using `terraform import`, import `aws_networkmanager_transit_gateway_peering` using the peering ID. For example:

```console
% terraform import aws_networkmanager_transit_gateway_peering.example peering-444555aaabbb11223
```

---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_peering"
description: |-
  Creates a peering connection between an AWS Cloud WAN core network and an AWS Transit Gateway.
---

# Resource: aws_networkmanager_transit_gateway_peering

Creates a peering connection between an AWS Cloud WAN core network and an AWS Transit Gateway.

## Example Usage

```terraform
resource "aws_networkmanager_transit_gateway_peering" "example" {
  core_network_id     = awscc_networkmanager_core_network.example.id
  transit_gateway_arn = aws_ec2_transit_gateway.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `core_network_id` - (Required) The ID of a core network.
* `tags` - (Optional) Key-value tags for the peering. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_arn` - (Required) The ARN of the transit gateway for the peering request.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Peering Amazon Resource Name (ARN).
* `core_network_arn` - The ARN of the core network.
* `edge_location` - The edge location for the peer.
* `id` - Peering ID.
* `owner_account_id` - The ID of the account owner.
* `peering_type` - The type of peering. This will be `TRANSIT_GATEWAY`.
* `resource_arn` - The resource ARN of the peer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `transit_gateway_peering_attachment_id` - The ID of the transit gateway peering attachment.

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

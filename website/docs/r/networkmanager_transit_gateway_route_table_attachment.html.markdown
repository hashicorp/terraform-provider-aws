---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_route_table_attachment"
description: |-
  Manages a Network Manager transit gateway route table attachment.
---

# Resource: aws_networkmanager_transit_gateway_route_table_attachment

Manages a Network Manager transit gateway route table attachment.

## Example Usage

```terraform
resource "aws_networkmanager_transit_gateway_route_table_attachment" "example" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.example.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.example.arn
}
```

## Argument Reference

The following arguments are required:

* `peering_id` - (Required) ID of the peer for the attachment.
* `transit_gateway_route_table_arn` - (Required) ARN of the transit gateway route table for the attachment.

The following arguments are optional:

* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Attachment ARN.
* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of the core network.
* `core_network_id` - ID of the core network.
* `edge_location` - Edge location for the peer.
* `id` - ID of the attachment.
* `owner_account_id` - ID of the attachment account owner.
* `resource_arn` - Attachment resource ARN.
* `segment_name` - Name of the segment attachment.
* `state` - State of the attachment.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_transit_gateway_route_table_attachment` using the attachment ID. For example:

```terraform
import {
  to = aws_networkmanager_transit_gateway_route_table_attachment.example
  id = "attachment-0f8fa60d2238d1bd8"
}
```

Using `terraform import`, import `aws_networkmanager_transit_gateway_route_table_attachment` using the attachment ID. For example:

```console
% terraform import aws_networkmanager_transit_gateway_route_table_attachment.example attachment-0f8fa60d2238d1bd8
```

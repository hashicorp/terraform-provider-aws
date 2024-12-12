---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_route_table_attachment"
description: |-
  Creates a transit gateway route table attachment.
---

# Resource: aws_networkmanager_transit_gateway_route_table_attachment

Creates a transit gateway route table attachment.

## Example Usage

```terraform
resource "aws_networkmanager_transit_gateway_route_table_attachment" "example" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.example.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `peering_id` - (Required) The ID of the peer for the attachment.
* `tags` - (Optional) Key-value tags for the attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transit_gateway_route_table_arn` - (Required) The ARN of the transit gateway route table for the attachment.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Attachment Amazon Resource Name (ARN).
* `attachment_policy_rule_number` - The policy rule number associated with the attachment.
* `attachment_type` - The type of attachment.
* `core_network_arn` - The ARN of the core network.
* `core_network_id` - The ID of the core network.
* `edge_location` - The edge location for the peer.
* `id` - The ID of the attachment.
* `owner_account_id` - The ID of the attachment account owner.
* `resource_arn` - The attachment resource ARN.
* `segment_name` - The name of the segment attachment.
* `state` - The state of the attachment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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

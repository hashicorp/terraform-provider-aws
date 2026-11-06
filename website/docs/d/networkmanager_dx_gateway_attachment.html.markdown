---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_dx_gateway_attachment"
description: |-
  Provides details about an existing AWS Network Manager Direct Connect gateway attachment.
---

# Data Source: aws_networkmanager_dx_gateway_attachment

Provides details about an existing AWS Network Manager Direct Connect gateway attachment.

When searching by `core_network_id`, the arguments must match exactly one Direct Connect gateway attachment. Use [`aws_networkmanager_attachments`](/docs/providers/aws/d/networkmanager_attachments.html) to find several attachments.

## Example Usage

### By Attachment ID

```terraform
data "aws_networkmanager_dx_gateway_attachment" "example" {
  attachment_id = "attachment-0123456789abcdef0"
}
```

### By Direct Connect Gateway

```terraform
data "aws_networkmanager_dx_gateway_attachment" "example" {
  core_network_id            = "core-network-0123456789abcdef0"
  direct_connect_gateway_arn = "arn:aws:directconnect::123456789012:dx-gateway/01234567-89ab-cdef-0123-456789abcdef"
}
```

### By Edge Locations and Tags

```terraform
data "aws_networkmanager_dx_gateway_attachment" "example" {
  core_network_id = "core-network-0123456789abcdef0"
  edge_locations  = ["us-east-1"]

  tags = {
    Name = "example-dx-gateway-attachment"
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `attachment_id` - (Optional) ID of the attachment. Exactly one of `attachment_id` or `core_network_id` must be set.
* `core_network_id` - (Optional) ID of the core network to search for the attachment. Exactly one of `attachment_id` or `core_network_id` must be set.
* `direct_connect_gateway_arn` - (Optional) ARN of the Direct Connect gateway. Conflicts with `attachment_id`.
* `edge_locations` - (Optional) AWS Regions of the core network edges that the attachment must be associated with. An attachment matches when it's associated with all of the given edges, and it may be associated with more. Conflicts with `attachment_id`.
* `state` - (Optional) State of the attachment. While its edge locations are being changed, the attachment is in the `UPDATING` state. Conflicts with `attachment_id`. Allowed values are: `AVAILABLE`, `CREATING`, `DELETING`, `FAILED`, `PENDING_ATTACHMENT_ACCEPTANCE`, `PENDING_NETWORK_UPDATE`, `PENDING_TAG_ACCEPTANCE`, `REJECTED`, and `UPDATING`.
* `tags` - (Optional) Map of tags that the attachment must have. An attachment matches when it has all of the given tags. Conflicts with `attachment_id`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the attachment.
* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of the core network.
* `created_at` - Timestamp when the attachment was created.
* `last_modification_errors` - Errors associated with the last attachment request. See [`last_modification_errors` Block](#last_modification_errors-block) below.
* `nat_gateway_support` - State of NAT gateway support on the attachment. Returned only to the core network owner.
* `nat_system_route_support` - Whether the core network installs system routes for the NAT CIDR blocks of the attachment's NAT rules. Returned only to the core network owner.
* `network_function_group_name` - Name of the network function group.
* `owner_account_id` - ID of the attachment account owner.
* `proposed_network_function_group_change` - Proposed change to the network function group associated with the attachment. See [`proposed_network_function_group_change` Block](#proposed_network_function_group_change-block) below.
* `proposed_segment_change` - Proposed move of the attachment from one segment to another. See [`proposed_segment_change` Block](#proposed_segment_change-block) below.
* `resource_arn` - ARN of the attached Direct Connect gateway.
* `routing_policy_label` - Routing policy label associated with the attachment.
* `segment_name` - Name of the segment the attachment is associated with.
* `updated_at` - Timestamp when the attachment was last updated.

### `last_modification_errors` Block

`last_modification_errors` supports:

* `code` - Error code.
* `message` - Error message.
* `request_id` - ID of the request.
* `resource_arn` - ARN of the requested attachment resource.

### `proposed_network_function_group_change` Block

`proposed_network_function_group_change` supports:

* `attachment_policy_rule_number` - Proposed new attachment policy rule number.
* `network_function_group_name` - Proposed name of the network function group.
* `tags` - Proposed tags.

### `proposed_segment_change` Block

`proposed_segment_change` supports:

* `attachment_policy_rule_number` - Proposed new attachment policy rule number.
* `segment_name` - Proposed name of the segment.
* `tags` - Proposed tags.

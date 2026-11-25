---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachments"
description: |-
  Provides details about existing AWS Network Manager core network attachments.
---

# Data Source: aws_networkmanager_attachments

Provides details about existing AWS Network Manager core network attachments.

## Example Usage

### Basic Usage

```terraform
data "aws_networkmanager_attachments" "example" {
  core_network_id = "core-network-0123456789abcdef0"
}
```

### Filter by Type, State, and Tags

```terraform
data "aws_networkmanager_attachments" "example" {
  core_network_id = "core-network-0123456789abcdef0"
  attachment_type = "VPC"
  state           = "AVAILABLE"

  tags = {
    Name = "example-vpc-attachment"
  }
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of the core network.

The following arguments are optional:

* `attachment_type` - (Optional) Type of attachment. Allowed values are: `CONNECT`, `DIRECT_CONNECT_GATEWAY`, `SITE_TO_SITE_VPN`, `TRANSIT_GATEWAY_ROUTE_TABLE`, and `VPC`.
* `edge_location` - (Optional) AWS Region where the edge is located. Direct Connect gateway attachments match when the Region is one of their `edge_locations`.
* `state` - (Optional) State of the attachment. Allowed values are: `AVAILABLE`, `CREATING`, `DELETING`, `FAILED`, `PENDING_ATTACHMENT_ACCEPTANCE`, `PENDING_NETWORK_UPDATE`, `PENDING_TAG_ACCEPTANCE`, `REJECTED`, and `UPDATING`.
* `tags` - (Optional) Map of tags that each attachment must have. An attachment matches when it has all of the given tags.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `attachments` - List of attachments. See [`attachments` Block](#attachments-block) below.
* `ids` - IDs of the attachments.

### `attachments` Block

`attachments` supports:

* `arn` - ARN of the attachment.
* `attachment_id` - ID of the attachment.
* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `attachment_type` - Type of attachment.
* `core_network_arn` - ARN of the core network.
* `core_network_id` - ID of the core network.
* `created_at` - Timestamp when the attachment was created.
* `edge_location` - AWS Region where the edge is located. Not set for Direct Connect gateway attachments.
* `edge_locations` - AWS Regions where the edges are located. Set only for Direct Connect gateway attachments.
* `last_modification_errors` - Errors associated with the last attachment request. See [`last_modification_errors` Block](#last_modification_errors-block) below.
* `nat_gateway_support` - State of NAT gateway support on the attachment. Returned only to the core network owner for attachment types that support NAT.
* `nat_system_route_support` - Whether the core network installs system routes for the NAT CIDR blocks of the attachment's NAT rules. Returned only to the core network owner for attachment types that support NAT.
* `network_function_group_name` - Name of the network function group.
* `owner_account_id` - ID of the attachment account owner.
* `proposed_network_function_group_change` - Proposed change to the network function group associated with the attachment. See [`proposed_network_function_group_change` Block](#proposed_network_function_group_change-block) below.
* `proposed_segment_change` - Proposed move of the attachment from one segment to another. See [`proposed_segment_change` Block](#proposed_segment_change-block) below.
* `resource_arn` - ARN of the attached resource.
* `segment_name` - Name of the segment the attachment is associated with.
* `state` - State of the attachment.
* `tags` - Map of tags assigned to the attachment.
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

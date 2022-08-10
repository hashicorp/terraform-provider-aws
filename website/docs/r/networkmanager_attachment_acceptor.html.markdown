---
subcategory: "NetworkManager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachment_acceptor"
description: |-
  Terraform resource for managing an AWS NetworkManager VpcAttachment.
---

# Resource: aws_networkmanager_attachment_acceptor

Terraform resource for managing an AWS NetworkManager Attachment Acceptor.

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_attachment_acceptor" "test" {
    attachment_id = aws_networkmanager_vpc_attachment.core_network.id
}
```

## Argument Reference

The following arguments are required:

* `attachment_id` - (Required) The ID of the attachment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `attachment_policy_rule_number` - The policy rule number associated with the attachment.
* `attachment_type` - The type of attachment.
* `core_network_arn` - The ARN of a core network.
* `core_network_id` - The id of a core network.
* `edge_location` - The Region where the edge is located.
* `owner_account_id` - The ID of the attachment account owner.
* `resource_arn` - The attachment resource ARN.
* `segment_name` - The name of the segment attachment.
* `state` - The state of the attachment.

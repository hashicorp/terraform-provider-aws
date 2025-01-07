---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachment_accepter"
description: |-
  Terraform resource for managing an AWS Network Manager Attachment Accepter.
---

# Resource: aws_networkmanager_attachment_accepter

Terraform resource for managing an AWS Network Manager Attachment Accepter.

## Example Usage

### Example with VPC attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.vpc.id
  attachment_type = aws_networkmanager_vpc_attachment.vpc.attachment_type
}
```

### Example with site-to-site VPN attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.vpn.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.vpn.attachment_type
}
```

## Argument Reference

The following arguments are required:

- `attachment_id` - (Required) The ID of the attachment.
- `attachment_type` - (Required) The type of attachment. Valid values can be found in the [AWS Documentation](https://docs.aws.amazon.com/networkmanager/latest/APIReference/API_ListAttachments.html#API_ListAttachments_RequestSyntax)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `attachment_policy_rule_number` - The policy rule number associated with the attachment.
- `core_network_arn` - The ARN of a core network.
- `core_network_id` - The id of a core network.
- `edge_location` - The Region where the edge is located. This is returned for all attachment types except a Direct Connect gateway attachment, which instead returns `edge_locations`.
- `edge_locations` - The edge locations that the Direct Connect gateway is associated with. This is returned only for Direct Connect gateway attachments. All other attachment types return `edge_location`
- `owner_account_id` - The ID of the attachment account owner.
- `resource_arn` - The attachment resource ARN.
- `segment_name` - The name of the segment attachment.
- `state` - The state of the attachment.

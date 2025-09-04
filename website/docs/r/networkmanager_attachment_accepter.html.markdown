---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachment_accepter"
description: |-
  Manages an AWS Network Manager Attachment Accepter.
---

# Resource: aws_networkmanager_attachment_accepter

Manages an AWS Network Manager Attachment Accepter.

Use this resource to accept cross-account attachments in AWS Network Manager. When an attachment is created in one account and needs to be accepted by another account that owns the core network, this resource handles the acceptance process.

## Example Usage

### VPC Attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_vpc_attachment.example.id
  attachment_type = aws_networkmanager_vpc_attachment.example.attachment_type
}
```

### Site-to-Site VPN Attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.example.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.example.attachment_type
}
```

### Connect Attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_connect_attachment.example.id
  attachment_type = aws_networkmanager_connect_attachment.example.attachment_type
}
```

### Transit Gateway Route Table Attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_transit_gateway_route_table_attachment.example.id
  attachment_type = aws_networkmanager_transit_gateway_route_table_attachment.example.attachment_type
}
```

### Direct Connect Gateway Attachment

```terraform
resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_dx_gateway_attachment.example.id
  attachment_type = aws_networkmanager_dx_gateway_attachment.example.attachment_type
}
```

## Argument Reference

The following arguments are required:

* `attachment_id` - (Required) ID of the attachment.
* `attachment_type` - (Required) Type of attachment. Valid values: `CONNECT`, `DIRECT_CONNECT_GATEWAY`, `SITE_TO_SITE_VPN`, `TRANSIT_GATEWAY_ROUTE_TABLE`, `VPC`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `attachment_policy_rule_number` - Policy rule number associated with the attachment.
* `core_network_arn` - ARN of the core network.
* `core_network_id` - ID of the core network.
* `edge_location` - Region where the edge is located. This is returned for all attachment types except Direct Connect gateway attachments, which instead return `edge_locations`.
* `edge_locations` - Edge locations that the Direct Connect gateway is associated with. This is returned only for Direct Connect gateway attachments. All other attachment types return `edge_location`.
* `owner_account_id` - ID of the attachment account owner.
* `resource_arn` - Attachment resource ARN.
* `segment_name` - Name of the segment attachment.
* `state` - State of the attachment.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)

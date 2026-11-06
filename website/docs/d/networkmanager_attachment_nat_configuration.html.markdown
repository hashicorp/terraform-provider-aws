---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachment_nat_configuration"
description: |-
  Provides details about the 1:1 NAT configuration of a Network Manager Cloud WAN core network attachment.
---

# Data Source: aws_networkmanager_attachment_nat_configuration

Provides details about the 1:1 NAT configuration of a Network Manager Cloud WAN core network attachment. For more information, see [1:1 NAT in AWS Cloud WAN](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-nat.html).

## Example Usage

### Basic Usage

```terraform
data "aws_networkmanager_attachment_nat_configuration" "example" {
  attachment_id = aws_networkmanager_vpc_attachment.example.id
}
```

## Argument Reference

The following arguments are required:

* `attachment_id` - (Required) ID of the attachment. Must be read with credentials for the core network owner account, including for attachments owned by other accounts.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `nat_gateway_rule` - NAT gateway rules for the attachment. Empty if the attachment has no rules. See [`nat_gateway_rule` Block](#nat_gateway_rule-block) below.
* `nat_gateway_support` - State of NAT gateway support for the attachment, such as `ENABLED` or `DISABLED`.
* `nat_system_route_support` - Whether the core network installs routes for the NAT CIDR blocks in the attachment's segment. See [Routing behavior for 1:1 NAT](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-nat-routing.html).

### `nat_gateway_rule` Block

The `nat_gateway_rule` block supports:

* `match_source_cidr` - Original IPv4 CIDR block on the attachment that is translated.
* `nat_cidr` - IPv4 CIDR block that `match_source_cidr` is translated to.
* `rule_number` - Number that identifies the rule.
* `state` - State of the rule.

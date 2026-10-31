---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_attachment_nat_configuration"
description: |-
  Manages the 1:1 NAT configuration of a Network Manager Cloud WAN core network attachment.
---

# Resource: aws_networkmanager_attachment_nat_configuration

Manages the 1:1 NAT configuration of a Network Manager Cloud WAN core network attachment, which statically translates IPv4 CIDR blocks on the attachment to NAT CIDR blocks so that networks with overlapping address space can communicate. For more information, see [1:1 NAT in AWS Cloud WAN](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-nat.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = [aws_subnet.example.arn]
  vpc_arn         = aws_vpc.example.arn
}

resource "aws_networkmanager_attachment_nat_configuration" "example" {
  attachment_id       = aws_networkmanager_vpc_attachment.example.id
  core_network_id     = aws_networkmanager_vpc_attachment.example.core_network_id
  nat_gateway_support = "ENABLE"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
```

### Cross-Account Attachment

The core network is shared with another account through AWS RAM. The other account creates the attachment, and the core network owner accepts it and then configures NAT.

```terraform
resource "aws_networkmanager_vpc_attachment" "example" {
  provider = aws.attachment_owner

  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = [aws_subnet.example.arn]
  vpc_arn         = aws_vpc.example.arn
}

resource "aws_networkmanager_attachment_accepter" "example" {
  attachment_id   = aws_networkmanager_vpc_attachment.example.id
  attachment_type = aws_networkmanager_vpc_attachment.example.attachment_type
}

resource "aws_networkmanager_attachment_nat_configuration" "example" {
  attachment_id       = aws_networkmanager_attachment_accepter.example.attachment_id
  core_network_id     = aws_networkmanager_attachment_accepter.example.core_network_id
  nat_gateway_support = "ENABLE"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
```

## Argument Reference

The following arguments are required:

* `attachment_id` - (Required, Forces new resource) ID of the attachment. For the supported attachment types, see [1:1 NAT in AWS Cloud WAN](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-nat.html). Must be managed with credentials for the core network owner account, including for attachments owned by other accounts.
* `core_network_id` - (Required, Forces new resource) ID of the core network that the attachment belongs to.
* `nat_gateway_rule` - (Required) NAT gateway rules for the attachment. At least one rule is required. For the maximum number of rules, see [Cloud WAN quotas](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-quotas.html). Every change replaces the attachment's full rule set. See [`nat_gateway_rule` Block](#nat_gateway_rule-block) below.
* `nat_gateway_support` - (Required) Whether NAT gateway support is enabled for the attachment. Valid values are `ENABLE` and `DISABLE`. With `DISABLE`, the rules are kept on the attachment but deactivated.

The following arguments are optional:

* `nat_system_route_support` - (Optional) Whether to install routes for the NAT CIDR blocks in the attachment's segment. Whether the value can be changed depends on the attachment type. See [Routing behavior for 1:1 NAT](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-nat-routing.html). If not set, the value AWS assigns for the attachment type is kept. When this resource is destroyed, the attachment keeps the last value set.

### `nat_gateway_rule` Block

The `nat_gateway_rule` block supports the following:

* `match_source_cidr` - (Required) Original IPv4 CIDR block on the attachment to translate.
* `nat_cidr` - (Required) IPv4 CIDR block that `match_source_cidr` is translated to. Must be in RFC 1918 or RFC 6598 address space, have the same prefix length as `match_source_cidr`, and not overlap other rule CIDR blocks on the attachment.
* `rule_number` - (Required) Number that identifies the rule. Rules aren't evaluated in order. Valid values are between `1` and `32766`. Must be unique within the resource.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networkmanager_attachment_nat_configuration.example
  identity = {
    attachment_id   = "attachment-0123456789abcdef0"
    core_network_id = "core-network-0fab1c1e1e1e1e1e1"
  }
}

resource "aws_networkmanager_attachment_nat_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `attachment_id` (String) ID of the attachment.
* `core_network_id` (String) ID of the core network.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Manager Attachment NAT Configurations using the core network ID and attachment ID separated by a comma (`,`). For example:

```terraform
import {
  to = aws_networkmanager_attachment_nat_configuration.example
  id = "core-network-0fab1c1e1e1e1e1e1,attachment-0123456789abcdef0"
}
```

Using `terraform import`, import Network Manager Attachment NAT Configurations using the core network ID and attachment ID separated by a comma (`,`). For example:

```console
% terraform import aws_networkmanager_attachment_nat_configuration.example core-network-0fab1c1e1e1e1e1e1,attachment-0123456789abcdef0
```

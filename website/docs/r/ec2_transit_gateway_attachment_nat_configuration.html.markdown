---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_attachment_nat_configuration"
description: |-
  Manages the 1:1 NAT configuration of an EC2 Transit Gateway attachment.
---

# Resource: aws_ec2_transit_gateway_attachment_nat_configuration

Manages the 1:1 NAT configuration of an EC2 Transit Gateway attachment, which statically translates IPv4 CIDR blocks on the attachment to NAT CIDR blocks so that networks with overlapping address space can communicate. For more information, see [1:1 NAT in AWS Transit Gateway](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-nat.html).

~> **Note:** On VPC attachments, enabling NAT replaces the propagated route for each VPC CIDR block that a rule exactly matches with a route to its NAT CIDR block. Other attachments that reach the VPC by its original addresses lose connectivity.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}

resource "aws_ec2_transit_gateway_attachment_nat_configuration" "example" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}
```

### Overlapping Spoke VPCs

Two spoke VPCs both use `10.0.0.0/16`. Each attachment translates the VPC's addresses to its own NAT CIDR block.

```terraform
resource "aws_ec2_transit_gateway_attachment_nat_configuration" "spoke_a" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.spoke_a.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.64.0.0/16"
  }
}

resource "aws_ec2_transit_gateway_attachment_nat_configuration" "spoke_b" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.spoke_b.id
  nat_gateway_support           = "enable"

  nat_gateway_rule {
    rule_number       = 100
    match_source_cidr = "10.0.0.0/16"
    nat_cidr          = "100.65.0.0/16"
  }
}
```

## Argument Reference

The following arguments are required:

* `nat_gateway_rule` - (Required) NAT gateway rules for the attachment. At least one rule is required. For the maximum number of rules, see [1:1 NAT quotas](https://docs.aws.amazon.com/vpc/latest/tgw/transit-gateway-quotas.html). Every change replaces the attachment's full rule set. See [`nat_gateway_rule` Block](#nat_gateway_rule-block) below.
* `nat_gateway_support` - (Required) Whether NAT gateway support is enabled for the attachment. Valid values are `enable` and `disable`. With `disable`, the rules are kept on the attachment but deactivated.
* `transit_gateway_attachment_id` - (Required, Forces new resource) ID of the transit gateway attachment. Must be managed with credentials for the account that owns the transit gateway, including for attachments owned by other accounts.

The following arguments are optional:

* `nat_system_route_support` - (Optional) Whether the transit gateway installs routes for the NAT CIDR blocks in the route tables to which the attachment propagates. Valid values are `enable` and `disable`. Whether the value can be changed depends on the attachment type. See [NAT system route support](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-nat-routing.html). Defaults to the attachment's current value. When this resource is destroyed, the attachment keeps the last value set.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `nat_gateway_rule` Block

No CIDR block in a rule, source or NAT, may overlap any other rule's CIDR blocks on the same attachment.

The `nat_gateway_rule` block supports the following:

* `match_source_cidr` - (Required) Original IPv4 CIDR block on the attachment to translate.
* `nat_cidr` - (Required) IPv4 CIDR block that `match_source_cidr` is translated to. Must be in RFC 1918 or RFC 6598 address space, have the same prefix length as `match_source_cidr`, and not overlap other rule CIDR blocks on the attachment.
* `rule_number` - (Required) Number that identifies the rule. Rules aren't evaluated in order. Valid values are between `1` and `32766`. Must be unique within the resource.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_attachment_nat_configuration.example
  identity = {
    transit_gateway_attachment_id = "tgw-attach-0123456789abcdef0"
  }
}

resource "aws_ec2_transit_gateway_attachment_nat_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `transit_gateway_attachment_id` (String) ID of the transit gateway attachment.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Transit Gateway Attachment NAT Configurations using the `transit_gateway_attachment_id`. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_attachment_nat_configuration.example
  id = "tgw-attach-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Transit Gateway Attachment NAT Configurations using the `transit_gateway_attachment_id`. For example:

```console
% terraform import aws_ec2_transit_gateway_attachment_nat_configuration.example tgw-attach-0123456789abcdef0
```

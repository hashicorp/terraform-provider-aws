---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_attachment_nat_configuration"
description: |-
  Provides details about the 1:1 NAT configuration of an EC2 Transit Gateway attachment.
---

# Data Source: aws_ec2_transit_gateway_attachment_nat_configuration

Provides details about the 1:1 NAT configuration of an EC2 Transit Gateway attachment. For more information, see [1:1 NAT in AWS Transit Gateway](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-nat.html).

## Example Usage

### Basic Usage

```terraform
data "aws_ec2_transit_gateway_attachment_nat_configuration" "example" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id
}
```

## Argument Reference

The following arguments are required:

* `transit_gateway_attachment_id` - (Required) ID of the transit gateway attachment. Must be read with credentials for the account that owns the transit gateway, including for attachments owned by other accounts.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `nat_gateway_rule` - NAT gateway rules for the attachment. Empty if the attachment has no rules. See [`nat_gateway_rule` Block](#nat_gateway_rule-block) below.
* `nat_gateway_support` - State of NAT gateway support for the attachment, such as `enabled` or `disabled`.
* `nat_system_route_support` - Whether the transit gateway installs routes for the NAT CIDR blocks in the route tables to which the attachment propagates. See [NAT system route support](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-nat-routing.html).

### `nat_gateway_rule` Block

The `nat_gateway_rule` block supports:

* `match_source_cidr` - Original IPv4 CIDR block on the attachment that is translated.
* `nat_cidr` - IPv4 CIDR block that `match_source_cidr` is translated to.
* `rule_number` - Number that identifies the rule.
* `state` - State of the rule.

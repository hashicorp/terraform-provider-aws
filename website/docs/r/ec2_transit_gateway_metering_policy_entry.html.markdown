---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_metering_policy_entry"
description: |-
  Manages an EC2 Transit Gateway Metering Policy Entry
---

# Resource: aws_ec2_transit_gateway_metering_policy_entry

Manages an EC2 Transit Gateway Metering Policy Entry. Each entry defines a traffic matching rule within a [Transit Gateway Metering Policy](ec2_transit_gateway_metering_policy.html) that determines which account is charged for matching traffic flows.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_metering_policy_entry" "example" {
  transit_gateway_metering_policy_id = aws_ec2_transit_gateway_metering_policy.example.transit_gateway_metering_policy_id
  policy_rule_number                 = 100
  metered_account                    = "source-attachment-owner"
}
```

### Full Traffic Matching Rule

```terraform
resource "aws_ec2_transit_gateway_metering_policy_entry" "example" {
  transit_gateway_metering_policy_id = aws_ec2_transit_gateway_metering_policy.example.transit_gateway_metering_policy_id
  policy_rule_number                 = 200
  metered_account                    = "destination-attachment-owner"
  source_cidr_block                  = "10.0.0.0/8"
  destination_cidr_block             = "172.16.0.0/12"
  protocol                           = "6"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `transit_gateway_metering_policy_id` - (Required, Forces new resource) EC2 Transit Gateway Metering Policy identifier.
* `policy_rule_number` - (Required, Forces new resource) Rule number for this entry. Lower numbers have higher priority. Valid values are between `1` and `32766`.
* `metered_account` - (Required, Forces new resource) The account to charge for matching traffic. Valid values are `source-attachment-owner` or `destination-attachment-owner`.
* `source_cidr_block` - (Optional, Forces new resource) Source CIDR block to match. If not specified, all source CIDR blocks are matched.
* `destination_cidr_block` - (Optional, Forces new resource) Destination CIDR block to match. If not specified, all destination CIDR blocks are matched.
* `protocol` - (Optional, Forces new resource) Protocol number to match (e.g., `6` for TCP, `17` for UDP). If not specified, all protocols are matched.
* `source_attachment_resource_type` - (Optional, Forces new resource) Source attachment resource type to match. Valid values are `vpc`, `vpn`, `direct-connect-gateway`, `connect`, `peering`, `tgw-peering`.
* `destination_attachment_resource_type` - (Optional, Forces new resource) Destination attachment resource type to match. Valid values are `vpc`, `vpn`, `direct-connect-gateway`, `connect`, `peering`, `tgw-peering`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_metering_policy_entry` using the composite identifier `{transit_gateway_metering_policy_id},{policy_rule_number}`. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_metering_policy_entry.example
  id = "tgw-policy-12345678,100"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_metering_policy_entry` using the composite identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_metering_policy_entry.example tgw-policy-12345678,100
```

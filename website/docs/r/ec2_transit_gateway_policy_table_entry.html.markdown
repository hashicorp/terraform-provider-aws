---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_policy_table_entry"
description: |-
  Manages an EC2 Transit Gateway Policy Table Entry
---

# Resource: aws_ec2_transit_gateway_policy_table_entry

Manages an EC2 Transit Gateway Policy Table Entry. Each entry defines a traffic matching rule within a [Transit Gateway Policy Table](ec2_transit_gateway_policy_table.html) that routes matching traffic to a specified transit gateway route table, enabling Policy-Based Routing (PBR).

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_policy_table_entry" "example" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.example.id
  policy_rule_number               = 100
  target_route_table_id            = aws_ec2_transit_gateway_route_table.example.id
}
```

### Full Traffic Matching Rule

```terraform
resource "aws_ec2_transit_gateway_policy_table_entry" "example" {
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.example.id
  policy_rule_number               = 200
  target_route_table_id            = aws_ec2_transit_gateway_route_table.example.id

  policy_rule {
    source_cidr_block       = "10.0.1.0/24"
    source_port_range       = "*"
    destination_cidr_block  = "10.0.2.0/24"
    destination_port_range  = "443"
    protocol                = "6"

    metadata {
      key   = "test"
      value = "test"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `policy_rule_number` - (Required, Forces new resource) Rule number for this entry. Lower numbers are evaluated first and take precedence. Enter an integer from 1 to 50,000. Leave gaps between numbers (for example, 100, 110, 120) so you can insert rules later without renumbering.
* `target_route_table_id` - (Required) ID of the transit gateway route table to use for traffic matching this rule.
* `transit_gateway_policy_table_id` - (Required, Forces new resource) EC2 Transit Gateway Policy Table identifier.

The following arguments are optional:

* `policy_rule` - (Optional) Matching criteria for the policy table entry. [See below](#policy_rule).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `policy_rule`

* `destination_cidr_block` - (Optional) Destination CIDR block to match. If not specified, all destination CIDR blocks are matched.
* `destination_port_range` - (Optional) Destination port or port range to match (e.g., `443` or `1024-65535`). Only valid when `protocol` is `6` (TCP) or `17` (UDP).
* `metadata` - (Optional) Metadata key/value tag associated with the policy rule. [See below](#metadata).
* `protocol` - (Optional) Protocol number to match (e.g., `6` for TCP, `17` for UDP). If not specified, all protocols are matched.
* `source_cidr_block` - (Optional) Source CIDR block to match. If not specified, all source CIDR blocks are matched.
* `source_port_range` - (Optional) Source port or port range to match (e.g., `443` or `1024-65535`). Only valid when `protocol` is `6` (TCP) or `17` (UDP).

### `metadata`

* `key` - (Optional) Metadata key name for the policy rule.
* `value` - (Optional) Metadata key value for the policy rule.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `state` - State of the transit gateway policy table entry.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_policy_table_entry.example
  identity = {
    transit_gateway_policy_table_id = "tgw-ptb-000000000fffffff"
    policy_rule_number              = "100"
  }
}

resource "aws_ec2_transit_gateway_policy_table_entry" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `transit_gateway_policy_table_id` (String) EC2 Transit Gateway Policy Table identifier.
* `policy_rule_number` (String) Rule number for this entry.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_policy_table_entry` using the composite identifier `{transit_gateway_policy_table_id},{policy_rule_number}`. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_policy_table_entry.example
  id = "tgw-ptb-000000000fffffff,100"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_policy_table_entry` using the composite identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_policy_table_entry.example tgw-ptb-000000000fffffff,100
```

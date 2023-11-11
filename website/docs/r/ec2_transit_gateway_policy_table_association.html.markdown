---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_policy_table_association"
description: |-
  Manages an EC2 Transit Gateway Policy Table association
---

# Resource: aws_ec2_transit_gateway_policy_table_association

Manages an EC2 Transit Gateway Policy Table association.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_policy_table_association" "example" {
  transit_gateway_attachment_id   = aws_networkmanager_transit_gateway_peering.example.transit_gateway_peering_attachment_id
  transit_gateway_policy_table_id = aws_ec2_transit_gateway_policy_table.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `transit_gateway_attachment_id` - (Required) Identifier of EC2 Transit Gateway Attachment.
* `transit_gateway_policy_table_id` - (Required) Identifier of EC2 Transit Gateway Policy Table.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - EC2 Transit Gateway Policy Table identifier combined with EC2 Transit Gateway Attachment identifier
* `resource_id` - Identifier of the resource
* `resource_type` - Type of the resource

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_policy_table_association` using the EC2 Transit Gateway Policy Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_policy_table_association.example
  id = "tgw-rtb-12345678_tgw-attach-87654321"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_policy_table_association` using the EC2 Transit Gateway Policy Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_policy_table_association.example tgw-rtb-12345678_tgw-attach-87654321
```

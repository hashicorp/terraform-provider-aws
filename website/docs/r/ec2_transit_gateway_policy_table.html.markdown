---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_policy_table"
description: |-
  Manages an EC2 Transit Gateway Policy Table
---

# Resource: aws_ec2_transit_gateway_policy_table

Manages an EC2 Transit Gateway Policy Table.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_policy_table" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id

  tags = {
    Name = "Example Policy Table"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `transit_gateway_id` - (Required) EC2 Transit Gateway identifier.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Policy Table. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - EC2 Transit Gateway Policy Table Amazon Resource Name (ARN).
* `id` - EC2 Transit Gateway Policy Table identifier.
* `state` - The state of the EC2 Transit Gateway Policy Table.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_transit_gateway_policy_table` using the EC2 Transit Gateway Policy Table identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_policy_table.example
  id = "tgw-rtb-12345678"
}
```

Using `terraform import`, import `aws_ec2_transit_gateway_policy_table` using the EC2 Transit Gateway Policy Table identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_policy_table.example tgw-rtb-12345678
```

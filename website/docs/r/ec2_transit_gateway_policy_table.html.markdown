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

The following arguments are supported:

* `transit_gateway_id` - (Required) EC2 Transit Gateway identifier.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Policy Table. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - EC2 Transit Gateway Policy Table Amazon Resource Name (ARN).
* `id` - EC2 Transit Gateway Policy Table identifier.
* `state` - The state of the EC2 Transit Gateway Policy Table.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`aws_ec2_transit_gateway_policy_table` can be imported by using the EC2 Transit Gateway Policy Table identifier, e.g.,

```
$ terraform import aws_ec2_transit_gateway_policy_table.example tgw-rtb-12345678
```

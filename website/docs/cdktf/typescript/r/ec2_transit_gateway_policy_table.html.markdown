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

* `transitGatewayId` - (Required) EC2 Transit Gateway identifier.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Policy Table. If configured with a provider [`defaultTags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - EC2 Transit Gateway Policy Table Amazon Resource Name (ARN).
* `id` - EC2 Transit Gateway Policy Table identifier.
* `state` - The state of the EC2 Transit Gateway Policy Table.
* `tagsAll` - A map of tags assigned to the resource, including those inherited from the provider [`defaultTags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`awsEc2TransitGatewayPolicyTable` can be imported by using the EC2 Transit Gateway Policy Table identifier, e.g.,

```
$ terraform import aws_ec2_transit_gateway_policy_table.example tgw-rtb-12345678
```

<!-- cache-key: cdktf-0.17.0-pre.15 input-6200610860d8045f7e9f0172434a88391558b460f129907e25f4790f7bb77806 -->
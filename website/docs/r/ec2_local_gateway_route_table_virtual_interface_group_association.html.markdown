---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route_table_virtual_interface_group_association"
description: |-
  Manages an EC2 Local Gateway Route Table Virtual Interface Group Association
---

# Resource: aws_ec2_local_gateway_route_table_virtual_interface_group_association

Manages an EC2 Local Gateway Route Table Virtual Interface Group Association. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-local-gateways.html).

## Example Usage

```terraform
data "aws_ec2_local_gateway_route_table" "example" {
  outpost_arn = "arn:aws:outposts:us-west-2:123456789012:outpost/op-1234567890abcdef"
}

data "aws_ec2_local_gateway_virtual_interface_group" "example" {
  local_gateway_id = data.aws_ec2_local_gateway_route_table.example.local_gateway_id
}

resource "aws_ec2_local_gateway_route_table_virtual_interface_group_association" "example" {
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.example.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.example.id

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `local_gateway_route_table_id` - (Required) Identifier of EC2 Local Gateway Route Table.
* `local_gateway_virtual_interface_group_id` - (Required) Identifier of EC2 Local Gateway Virtual Interface Group.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of EC2 Local Gateway Route Table Virtual Interface Group Association.
* `local_gateway_id` - Identifier of the EC2 Local Gateway.
* `local_gateway_route_table_arn` - Amazon Resource Name (ARN) of the EC2 Local Gateway Route Table.
* `owner_id` - Identifier of the AWS account that owns the EC2 Local Gateway Virtual Interface Group Association.
* `state` - State of the EC2 Local Gateway Route Table Virtual Interface Group Association.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_local_gateway_route_table_virtual_interface_group_association` using the Local Gateway Route Table Virtual Interface Group Association identifier. For example:

```terraform
import {
  to = aws_ec2_local_gateway_route_table_virtual_interface_group_association.example
  id = "lgw-vif-grp-assoc-1234567890abcdef"
}
```

Using `terraform import`, import `aws_ec2_local_gateway_route_table_virtual_interface_group_association` using the Local Gateway Route Table Virtual Interface Group Association identifier. For example:

```console
% terraform import aws_ec2_local_gateway_route_table_virtual_interface_group_association.example lgw-vif-grp-assoc-1234567890abcdef
```

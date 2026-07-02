---
subcategory: "Outposts (EC2)"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route_table"
description: |-
  Manages an EC2 Local Gateway Route Table
---

# Resource: aws_ec2_local_gateway_route_table

Manages an EC2 Local Gateway Route Table. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-local-gateways.html#route-tables).

## Example Usage

```terraform
data "aws_ec2_local_gateway" "example" {
  id = "lgw-1234567890abcdef"
}

resource "aws_ec2_local_gateway_route_table" "example" {
  local_gateway_id = data.aws_ec2_local_gateway.example.id
  mode             = "direct-vpc-routing"
}
```

## Argument Reference

The following arguments are required:

* `local_gateway_id` - (Required) Identifier of the EC2 Local Gateway.
* `mode` - (Required) Mode of the Local Gateway Route Table. Valid values: `direct-vpc-routing`, `coip`.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Local Gateway Route Table.
* `local_gateway_route_table_id` - Identifier of the Local Gateway Route Table.
* `outpost_arn` - ARN of the Outpost.
* `owner_id` - AWS account identifier that owns the Local Gateway Route Table.
* `state` - State of the Local Gateway Route Table.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_ec2_local_gateway_route_table` using the Local Gateway Route Table identifier. For example:

```terraform
import {
  to = aws_ec2_local_gateway_route_table.example
  id = "lgw-rtb-1234567890abcdef"
}
```

Using `terraform import`, import `aws_ec2_local_gateway_route_table` using the Local Gateway Route Table identifier. For example:

```console
% terraform import aws_ec2_local_gateway_route_table.example lgw-rtb-1234567890abcdef
```

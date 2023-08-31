---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_vpc_link"
description: |-
  Manages an Amazon API Gateway Version 2 VPC Link.
---

# Resource: aws_apigatewayv2_vpc_link

Manages an Amazon API Gateway Version 2 VPC Link.

-> **Note:** Amazon API Gateway Version 2 VPC Links enable private integrations that connect HTTP APIs to private resources in a VPC.
To enable private integration for REST APIs, use the Amazon API Gateway Version 1 VPC Link [resource](/docs/providers/aws/r/api_gateway_vpc_link.html).

## Example Usage

```terraform
resource "aws_apigatewayv2_vpc_link" "example" {
  name               = "example"
  security_group_ids = [data.aws_security_group.example.id]
  subnet_ids         = data.aws_subnets.example.ids

  tags = {
    Usage = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name of the VPC Link. Must be between 1 and 128 characters in length.
* `security_group_ids` - (Required) Security group IDs for the VPC Link.
* `subnet_ids` - (Required) Subnet IDs for the VPC Link.
* `tags` - (Optional) Map of tags to assign to the VPC Link. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - VPC Link identifier.
* `arn` - VPC Link ARN.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_apigatewayv2_vpc_link` using the VPC Link identifier. For example:

```terraform
import {
  to = aws_apigatewayv2_vpc_link.example
  id = "aabbccddee"
}
```

Using `terraform import`, import `aws_apigatewayv2_vpc_link` using the VPC Link identifier. For example:

```console
% terraform import aws_apigatewayv2_vpc_link.example aabbccddee
```

---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_vpc_link"
description: |-
  Provides an API Gateway VPC Link.
---

# Resource: aws_api_gateway_vpc_link

Provides an API Gateway VPC Link.

-> **Note:** Amazon API Gateway Version 1 VPC Links enable private integrations that connect REST APIs to private resources in a VPC.
To enable private integration for HTTP APIs, use the Amazon API Gateway Version 2 VPC Link [resource](/docs/providers/aws/r/apigatewayv2_vpc_link.html).

## Example Usage

```terraform
resource "aws_lb" "example" {
  name               = "example"
  internal           = true
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id = "12345"
  }
}

resource "aws_api_gateway_vpc_link" "example" {
  name        = "example"
  description = "example description"
  target_arns = [aws_lb.example.arn]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name used to label and identify the VPC link.
* `description` - (Optional) Description of the VPC link.
* `target_arns` - (Required, ForceNew) List of network load balancer arns in the VPC targeted by the VPC link. Currently AWS only supports 1 target.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the VpcLink.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway VPC Link using the `id`. For example:

```terraform
import {
  to = aws_api_gateway_vpc_link.example
  id = "12345abcde"
}
```

Using `terraform import`, import API Gateway VPC Link using the `id`. For example:

```console
% terraform import aws_api_gateway_vpc_link.example 12345abcde
```

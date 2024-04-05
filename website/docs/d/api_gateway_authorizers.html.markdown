---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_authorizers"
description: |-
  Provides details about multiple API Gateway Authorizers.
---

# Data Source: aws_api_gateway_authorizers

Provides details about multiple API Gateway Authorizers.

## Example Usage

```terraform
data "aws_api_gateway_authorizers" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
}
```

## Argument Reference

The following arguments are required:

* `rest_api_id` - (Required) ID of the associated REST API.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `ids` - List of Authorizer identifiers.

---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_apis"
description: |-
  Provides details about multiple Amazon API Gateway APIs.
---

# Data Source: aws_apigateway_apis

Provides details about multiple Amazon API Gateway Version 1 APIs.

## Example Usage

```terraform
data "aws_api_gateway_rest_apis" "example" {
}
```

## Attributes Reference

The following attributes are exported:

* `names` - Set of API names.

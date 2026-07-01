---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_integration"
description: |-
  Lists API Gateway V2 Integration resources.
---

# List Resource: aws_apigatewayv2_integration

Lists API Gateway V2 Integration resources.

## Example Usage

```terraform
list "aws_apigatewayv2_integration" "example" {
  provider = aws

  config {
    api_id = aws_apigatewayv2_api.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `api_id` - (Required) API identifier.
* `region` - (Optional) Region to query. Defaults to provider region.
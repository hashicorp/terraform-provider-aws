---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_portal"
description: |-
  Lists API Gateway Portal resources.
---

# List Resource: aws_apigatewayv2_portal

Lists API Gateway Portal resources.

## Example Usage

```terraform
list "aws_apigatewayv2_portal" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.

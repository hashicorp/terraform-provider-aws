---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_api"
description: |-
  Lists API Gateway REST API resources.
---

# List Resource: aws_api_gateway_rest_api

Lists API Gateway REST API resources.

## Example Usage

```terraform
list "aws_api_gateway_rest_api" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.

---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_api"
description: |-
  Lists API Gateway V2 API resources.
---

# List Resource: aws_apigatewayv2_api

Lists API Gateway V2 API resources.

## Example Usage

```terraform
list "aws_apigatewayv2_api" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.

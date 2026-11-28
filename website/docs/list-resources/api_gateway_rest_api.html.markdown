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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.

---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_deployment"
description: |-
  Lists API Gateway Deployment resources.
---

# List Resource: aws_api_gateway_deployment

Lists API Gateway Deployment resources for a specific REST API.

## Example Usage

```terraform
list "aws_api_gateway_deployment" "example" {
  provider = aws

  config {
    rest_api_id = aws_api_gateway_rest_api.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `rest_api_id` - (Required) ID of the associated REST API.

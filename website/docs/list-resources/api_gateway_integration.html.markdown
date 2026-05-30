---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_integration"
description: |-
  Lists API Gateway Integration resources.
---

# List Resource: aws_api_gateway_integration

Lists API Gateway Integration resources for a specific API Gateway Resource.

## Example Usage

```terraform
list "aws_api_gateway_integration" "example" {
  provider = aws

  config {
    rest_api_id = aws_api_gateway_rest_api.example.id
    resource_id = aws_api_gateway_resource.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `resource_id` - (Required) ID of the API Gateway Resource to list integrations from.
* `rest_api_id` - (Required) ID of the associated REST API.

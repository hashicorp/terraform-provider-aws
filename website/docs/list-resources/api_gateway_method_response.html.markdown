---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_method_response"
description: |-
  Lists API Gateway Method Response resources.
---

# List Resource: aws_api_gateway_method_response

Lists API Gateway Method Response resources for a specific API Gateway Method.

## Example Usage

```terraform
list "aws_api_gateway_method_response" "example" {
  provider = aws

  config {
    rest_api_id = aws_api_gateway_rest_api.example.id
    resource_id = aws_api_gateway_resource.example.id
    http_method = aws_api_gateway_method.example.http_method
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `http_method` - (Required) HTTP method of the API Gateway Method.
* `region` - (Optional) Region to query. Defaults to provider region.
* `resource_id` - (Required) ID of the API Gateway Resource.
* `rest_api_id` - (Required) ID of the associated REST API.

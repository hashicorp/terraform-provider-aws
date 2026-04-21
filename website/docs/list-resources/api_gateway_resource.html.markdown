---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_resource"
description: |-
  Lists API Gateway Resource resources.
---

# List Resource: aws_api_gateway_resource

Lists API Gateway Resource resources for a specific REST API.

## Example Usage

```terraform
list "aws_api_gateway_resource" "example" {
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

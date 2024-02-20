---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_model"
description: |-
  Provides a Model for a REST API Gateway.
---

# Resource: aws_api_gateway_model

Provides a Model for a REST API Gateway.

## Example Usage

```terraform
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_model" "MyDemoModel" {
  rest_api_id  = aws_api_gateway_rest_api.MyDemoAPI.id
  name         = "user"
  description  = "a JSON schema"
  content_type = "application/json"

  schema = jsonencode({
    type = "object"
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `rest_api_id` - (Required) ID of the associated REST API
* `name` - (Required) Name of the model
* `description` - (Optional) Description of the model
* `content_type` - (Required) Content type of the model
* `schema` - (Required) Schema of the model in a JSON form

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the model

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_model` using `REST-API-ID/NAME`. For example:

```terraform
import {
  to = aws_api_gateway_model.example
  id = "12345abcde/example"
}
```

Using `terraform import`, import `aws_api_gateway_model` using `REST-API-ID/NAME`. For example:

```console
% terraform import aws_api_gateway_model.example 12345abcde/example
```

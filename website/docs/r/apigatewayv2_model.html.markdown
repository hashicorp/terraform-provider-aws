---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_model"
description: |-
  Manages an Amazon API Gateway Version 2 model.
---

# Resource: aws_apigatewayv2_model

Manages an Amazon API Gateway Version 2 [model](https://docs.aws.amazon.com/apigateway/latest/developerguide/models-mappings.html#models-mappings-models).

## Example Usage

### Basic

```terraform
resource "aws_apigatewayv2_model" "example" {
  api_id       = aws_apigatewayv2_api.example.id
  content_type = "application/json"
  name         = "example"

  schema = jsonencode({
    "$schema" = "http://json-schema.org/draft-04/schema#"
    title     = "ExampleModel"
    type      = "object"

    properties = {
      id = {
        type = "string"
      }
    }
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) API identifier.
* `content_type` - (Required)  The content-type for the model, for example, `application/json`. Must be between 1 and 256 characters in length.
* `name` - (Required) Name of the model. Must be alphanumeric. Must be between 1 and 128 characters in length.
* `schema` - (Required) Schema for the model. This should be a [JSON schema draft 4](https://tools.ietf.org/html/draft-zyp-json-schema-04) model. Must be less than or equal to 32768 characters in length.
* `description` - (Optional) Description of the model. Must be between 1 and 128 characters in length.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Model identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_apigatewayv2_model` using the API identifier and model identifier. For example:

```terraform
import {
  to = aws_apigatewayv2_model.example
  id = "aabbccddee/1122334"
}
```

Using `terraform import`, import `aws_apigatewayv2_model` using the API identifier and model identifier. For example:

```console
% terraform import aws_apigatewayv2_model.example aabbccddee/1122334
```

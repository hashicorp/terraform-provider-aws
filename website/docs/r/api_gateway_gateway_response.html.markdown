---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_gateway_response"
description: |-
  Provides an API Gateway Gateway Response for a REST API Gateway.
---

# Resource: aws_api_gateway_gateway_response

Provides an API Gateway Gateway Response for a REST API Gateway.

## Example Usage

```terraform
resource "aws_api_gateway_rest_api" "main" {
  name = "MyDemoAPI"
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  status_code   = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{\"message\":$context.error.messageString}"
  }

  response_parameters = {
    "gatewayresponse.header.Authorization" = "'Basic'"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `rest_api_id` - (Required) String identifier of the associated REST API.
* `response_type` - (Required) Response type of the associated GatewayResponse.
* `status_code` - (Optional) HTTP status code of the Gateway Response.
* `response_templates` - (Optional) Map of templates used to transform the response body.
* `response_parameters` - (Optional) Map of parameters (paths, query strings and headers) of the Gateway Response.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_gateway_response` using `REST-API-ID/RESPONSE-TYPE`. For example:

```terraform
import {
  to = aws_api_gateway_gateway_response.example
  id = "12345abcde/UNAUTHORIZED"
}
```

Using `terraform import`, import `aws_api_gateway_gateway_response` using `REST-API-ID/RESPONSE-TYPE`. For example:

```console
% terraform import aws_api_gateway_gateway_response.example 12345abcde/UNAUTHORIZED
```

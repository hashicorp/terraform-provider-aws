---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_method_response"
description: |-
  Provides an HTTP Method Response for an API Gateway Resource.
---

# Resource: aws_api_gateway_method_response

Provides an HTTP Method Response for an API Gateway Resource.

## Example Usage

```terraform
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_resource" "MyDemoResource" {
  rest_api_id = aws_api_gateway_rest_api.MyDemoAPI.id
  parent_id   = aws_api_gateway_rest_api.MyDemoAPI.root_resource_id
  path_part   = "mydemoresource"
}

resource "aws_api_gateway_method" "MyDemoMethod" {
  rest_api_id   = aws_api_gateway_rest_api.MyDemoAPI.id
  resource_id   = aws_api_gateway_resource.MyDemoResource.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "MyDemoIntegration" {
  rest_api_id = aws_api_gateway_rest_api.MyDemoAPI.id
  resource_id = aws_api_gateway_resource.MyDemoResource.id
  http_method = aws_api_gateway_method.MyDemoMethod.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_method_response" "response_200" {
  rest_api_id = aws_api_gateway_rest_api.MyDemoAPI.id
  resource_id = aws_api_gateway_resource.MyDemoResource.id
  http_method = aws_api_gateway_method.MyDemoMethod.http_method
  status_code = "200"
}
```

## Argument Reference

This resource supports the following arguments:

* `rest_api_id` - (Required) ID of the associated REST API
* `resource_id` - (Required) API resource ID
* `http_method` - (Required) HTTP Method (`GET`, `POST`, `PUT`, `DELETE`, `HEAD`, `OPTIONS`, `ANY`)
* `status_code` - (Required) HTTP status code
* `response_models` - (Optional) Map of the API models used for the response's content type
* `response_parameters` - (Optional) Map of response parameters that can be sent to the caller.
   For example: `response_parameters = { "method.response.header.X-Some-Header" = true }`
   would define that the header `X-Some-Header` can be provided on the response.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_method_response` using `REST-API-ID/RESOURCE-ID/HTTP-METHOD/STATUS-CODE`. For example:

```terraform
import {
  to = aws_api_gateway_method_response.example
  id = "12345abcde/67890fghij/GET/200"
}
```

Using `terraform import`, import `aws_api_gateway_method_response` using `REST-API-ID/RESOURCE-ID/HTTP-METHOD/STATUS-CODE`. For example:

```console
% terraform import aws_api_gateway_method_response.example 12345abcde/67890fghij/GET/200
```

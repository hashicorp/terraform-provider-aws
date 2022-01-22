---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_export"
description: |-
  Get information on an API Gateway REST API Key
---

# Data Source: aws_api_gateway_export

## Example Usage

```terraform
data "aws_api_gateway_export" "example" {
  rest_api_id = aws_api_gateway_stage.example.rest_api_id
  stage_name  = aws_api_gateway_stage.example.stage_name
  export_type = "oas30"
}
```

## Argument Reference

* `export_type` - (Required) The type of export. Acceptable values are `oas30` for OpenAPI 3.0.x and `swagger` for Swagger/OpenAPI 2.0.
* `rest_api_id` - (Required) The identifier of the associated REST API.
* `stage_name` - (Required) The name of the Stage that will be exported.
* `accepts` - (Optional) The content-type of the export. Valid values are `application/json` and `application/yaml` are supported for `export_type` `ofoas30` and `swagger`.
* `parameters` - (Optional) A key-value map of query string parameters that specify properties of the export. the following parameters are supported: `extensions='integrations'` or `extensions='apigateway'` will export the API with x-amazon-apigateway-integration extensions. `extensions='authorizers'` will export the API with x-amazon-apigateway-authorizer extensions.

## Attributes Reference

* `id` - The `REST-API-ID:STAGE-NAME`
* `body` - The API Spec.
* `content_type` - The content-type header value in the HTTP response.
* `content_disposition` - The content-disposition header value in the HTTP response.

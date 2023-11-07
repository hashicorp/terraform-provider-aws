---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_export"
description: |-
  Exports a definition of an API in a particular output format and specification.
---

# Data Source: aws_apigatewayv2_export

Exports a definition of an API in a particular output format and specification.

## Example Usage

```terraform
data "aws_apigatewayv2_export" "test" {
  api_id        = aws_apigatewayv2_route.test.api_id
  specification = "OAS30"
  output_type   = "JSON"
}
```

## Argument Reference

This data source supports the following arguments:

* `api_id` - (Required) API identifier.
* `specification` - (Required) Version of the API specification to use. `OAS30`, for OpenAPI 3.0, is the only supported value.
* `output_type` - (Required) Output type of the exported definition file. Valid values are `JSON` and `YAML`.
* `export_version` - (Optional) Version of the API Gateway export algorithm. API Gateway uses the latest version by default. Currently, the only supported version is `1.0`.
* `include_extensions` - (Optional) Whether to include API Gateway extensions in the exported API definition. API Gateway extensions are included by default.
* `stage_name` - (Optional) Name of the API stage to export. If you don't specify this property, a representation of the latest API configuration is exported.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - API identifier.
* `body` - ID of the API.

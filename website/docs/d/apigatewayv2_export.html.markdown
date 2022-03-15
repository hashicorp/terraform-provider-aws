---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
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

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `specification` - (Required) The version of the API specification to use. `OAS30`, for OpenAPI 3.0, is the only supported value.
* `output_type` - (Required) The output type of the exported definition file. Valid values are `JSON` and `YAML`.
* `export_version` - (Optional) The version of the API Gateway export algorithm. API Gateway uses the latest version by default. Currently, the only supported version is `1.0`.
* `include_extensions` - (Optional) Specifies whether to include API Gateway extensions in the exported API definition. API Gateway extensions are included by default.
* `stage_name` - (Optional) The name of the API stage to export. If you don't specify this property, a representation of the latest API configuration is exported.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The API identifier.
* `body` - The id of the API.

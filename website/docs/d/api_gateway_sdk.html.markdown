---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_sdk"
description: |-
  Gets an API Gateway client SDK
---

# Data Source: aws_api_gateway_sdk

## Example Usage

```terraform
data "aws_api_gateway_sdk" "example" {
  rest_api_id = aws_api_gateway_stage.example.rest_api_id
  stage_name  = aws_api_gateway_stage.example.stage_name
  sdk_type    = "android"

  parameters = {
    groupId         = "example"
    artifactId      = "example"
    artifactVersion = "example"
    invokerPackage  = "example"
  }
}
```

## Argument Reference

* `rest_api_id` - (Required) Identifier of the associated REST API.
* `stage_name` - (Required) Name of the Stage that will be exported.
* `sdk_type` - (Required) Language for the generated SDK. Currently `java`, `javascript`, `android`, `objectivec` (for iOS), `swift` (for iOS), and `ruby` are supported.
* `parameters` - (Optional) Key-value map of query string parameters `sdk_type` properties of the SDK. For SDK Type of `objectivec` or `swift`, a parameter named `classPrefix` is required. For SDK Type of `android`, parameters named `groupId`, `artifactId`, `artifactVersion`, and `invokerPackage` are required. For SDK Type of `java`, parameters named `serviceName` and `javaPackageName` are required.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The `REST-API-ID:STAGE-NAME`
* `body` - SDK as a string.
* `content_type` - Content-type header value in the HTTP response.
* `content_disposition` - Content-disposition header value in the HTTP response.

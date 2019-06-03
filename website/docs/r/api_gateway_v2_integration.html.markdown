---
layout: "aws"
page_title: "AWS: aws_api_gateway_v2_integration"
sidebar_current: "docs-aws-resource-api-gateway-v2-integration"
description: |-
  Manages an Amazon API Gateway Version 2 integration.
---

# Resource: aws_api_gateway_v2_integration

Manages an Amazon API Gateway Version 2 integration.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_v2_integration" "example" {
  api_id           = "${aws_api_gateway_v2_api.example.id}"
  integration_type = "MOCK"
}
```

### Lambda Integration

```hcl
resource "aws_lambda_function" "example" {
  filename      = "example.zip"
  function_name = "Example"
  role          = "${aws_iam_role.example.arn}"
  handler       = "index.handler"
  runtime       = "nodejs10.x"
}

resource "aws_api_gateway_v2_integration" "example" {
  api_id           = "${aws_api_gateway_v2_api.example.id}"
  integration_type = "AWS"

  connection_type               = "INTERNET"
  content_handling_strategy     = "CONVERT_TO_TEXT"
  description                   = "Lambda example"
  integration_method            = "POST"
  integration_uri               = "${aws_lambda_function.example.invoke_arn}"
  passthrough_behavior          = "WHEN_NO_MATCH"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `integration_type` - (Required) The integration type of an integration.
Valid values: `AWS`, `AWS_PROXY`, `HTTP`, `HTTP_PROXY`, `MOCK`.
* `connection_id` - (Optional) The connection ID.
* `connection_type` - (Optional) The type of the network connection to the integration endpoint. Valid values: `INTERNET`, `VPC_LINK`. Default is `INTERNET`.
* `content_handling_strategy` - (Optional) How to handle response payload content type conversions. Valid values: `CONVERT_TO_BINARY`, `CONVERT_TO_TEXT`.
* `credentials_arn` - (Optional) The credentials required for the integration, if any.
* `description` - (Optional) The description of the integration.
* `integration_method` - (Optional) The integration's HTTP method. Must be specified if `integration_type` is not `MOCK`.
* `integration_uri` - (Optional) The URI of the Lambda function for a Lambda proxy integration, where `integration_type` is `AWS_PROXY`.
* `passthrough_behavior` - (Optional) The pass-through behavior for incoming requests based on the Content-Type header in the request, and the available mapping templates specified as the `request_templates` attribute. Valid values: `WHEN_NO_MATCH`, `WHEN_NO_TEMPLATES`, `NEVER`. Default is `WHEN_NO_MATCH`.
* `request_templates` - (Optional) A map of Velocity templates that are applied on the request payload based on the value of the Content-Type header sent by the client.
* `template_selection_expression` - (Optional) The [template selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-template-selection-expressions) for the integration.
* `timeout_milliseconds` - (Optional) Custom timeout between 50 and 29,000 milliseconds. The default value is 29,000 milliseconds or 29 seconds.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The integration identifier.
* `integration_response_selection_expression` - The [integration response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-integration-response-selection-expressions) for the integration.

## Import

`aws_api_gateway_v2_integration` can be imported by using the API identifier and integration identifier, e.g.

```
$ terraform import aws_api_gateway_v2_integration.example aabbccddee/1122334
```

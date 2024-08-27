---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_authorizer"
description: |-
  Manages an Amazon API Gateway Version 2 authorizer.
---

# Resource: aws_apigatewayv2_authorizer

Manages an Amazon API Gateway Version 2 authorizer.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic WebSocket API

```terraform
resource "aws_apigatewayv2_authorizer" "example" {
  api_id           = aws_apigatewayv2_api.example.id
  authorizer_type  = "REQUEST"
  authorizer_uri   = aws_lambda_function.example.invoke_arn
  identity_sources = ["route.request.header.Auth"]
  name             = "example-authorizer"
}
```

### Basic HTTP API

```terraform
resource "aws_apigatewayv2_authorizer" "example" {
  api_id                            = aws_apigatewayv2_api.example.id
  authorizer_type                   = "REQUEST"
  authorizer_uri                    = aws_lambda_function.example.invoke_arn
  identity_sources                  = ["$request.header.Authorization"]
  name                              = "example-authorizer"
  authorizer_payload_format_version = "2.0"
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) API identifier.
* `authorizer_type` - (Required) Authorizer type. Valid values: `JWT`, `REQUEST`.
Specify `REQUEST` for a Lambda function using incoming request parameters.
For HTTP APIs, specify `JWT` to use JSON Web Tokens.
* `name` - (Required) Name of the authorizer. Must be between 1 and 128 characters in length.
* `authorizer_credentials_arn` - (Optional) Required credentials as an IAM role for API Gateway to invoke the authorizer.
Supported only for `REQUEST` authorizers.
* `authorizer_payload_format_version` - (Optional) Format of the payload sent to an HTTP API Lambda authorizer. Required for HTTP API Lambda authorizers.
Valid values: `1.0`, `2.0`.
* `authorizer_result_ttl_in_seconds` - (Optional) Time to live (TTL) for cached authorizer results, in seconds. If it equals 0, authorization caching is disabled.
If it is greater than 0, API Gateway caches authorizer responses. The maximum value is 3600, or 1 hour. Defaults to `300`.
Supported only for HTTP API Lambda authorizers.
* `authorizer_uri` - (Optional) Authorizer's Uniform Resource Identifier (URI).
For `REQUEST` authorizers this must be a well-formed Lambda function URI, such as the `invoke_arn` attribute of the [`aws_lambda_function`](/docs/providers/aws/r/lambda_function.html) resource.
Supported only for `REQUEST` authorizers. Must be between 1 and 2048 characters in length.
* `enable_simple_responses` - (Optional) Whether a Lambda authorizer returns a response in a simple format. If enabled, the Lambda authorizer can return a boolean value instead of an IAM policy.
Supported only for HTTP APIs.
* `identity_sources` - (Optional) Identity sources for which authorization is requested.
For `REQUEST` authorizers the value is a list of one or more mapping expressions of the specified request parameters.
For `JWT` authorizers the single entry specifies where to extract the JSON Web Token (JWT) from inbound requests.
* `jwt_configuration` - (Optional) Configuration of a JWT authorizer. Required for the `JWT` authorizer type.
Supported only for HTTP APIs.

The `jwt_configuration` object supports the following:

* `audience` - (Optional) List of the intended recipients of the JWT. A valid JWT must provide an aud that matches at least one entry in this list.
* `issuer` - (Optional) Base domain of the identity provider that issues JSON Web Tokens, such as the `endpoint` attribute of the [`aws_cognito_user_pool`](/docs/providers/aws/r/cognito_user_pool.html) resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Authorizer identifier.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_apigatewayv2_authorizer` using the API identifier and authorizer identifier. For example:

```terraform
import {
  to = aws_apigatewayv2_authorizer.example
  id = "aabbccddee/1122334"
}
```

Using `terraform import`, import `aws_apigatewayv2_authorizer` using the API identifier and authorizer identifier. For example:

```console
% terraform import aws_apigatewayv2_authorizer.example aabbccddee/1122334
```

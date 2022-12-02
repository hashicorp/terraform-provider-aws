---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_authorizers"
description: |-
  Terraform data source for managing an AWS API Gateway authorizer.
---

# Data Source: aws_api_gateway_authorizers

Terraform data source for managing an AWS API Gateway authorizer.

## Example Usage

### Basic Usage

```terraform
data "aws_api_gateway_authorizers" "example" {
  rest_api_id = aws_api_gateway_rest_api.test.id
}
```

## Argument Reference

The following arguments are required:

* `rest_api_id` - (Required) REST API id that owns the resource. If no REST API is found, an error will be returned.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `items` (List of Object).

### Nested Attributes for `items`

* `name` - Name of the REST API Authorizer to look up. If no REST API Authorizer is found with this name, an error will be returned.
* `authorizer_uri` - Authorizer's Uniform Resource Identifier (URI). This must be a well-formed Lambda function URI in the form of `arn:aws:apigateway:{region}:lambda:path/{service_api}`,
  e.g., `arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-west-2:012345678912:function:my-function/invocations`
* `identity_source` - Source of the identity in an incoming request. Defaults to `method.request.header.Authorization`. For `REQUEST` type, this may be a comma-separated list of values, including headers, query string parameters and stage variables - e.g., `"method.request.header.SomeHeaderName,method.request.querystring.SomeQueryStringName,stageVariables.SomeStageVariableName"`
* `type` - Type of the authorizer. Possible values are `TOKEN` for a Lambda function using a single authorization token submitted in a custom header, `REQUEST` for a Lambda function using incoming request parameters, or `COGNITO_USER_POOLS` for using an Amazon Cognito user pool. Defaults to `TOKEN`.
* `authorizer_result_ttl_in_seconds` - TTL of cached authorizer results in seconds. Defaults to `300`.

---
layout: "aws"
page_title: "AWS: aws_api_gateway_v2_authorizer"
sidebar_current: "docs-aws-resource-api-gateway-v2-authorizer"
description: |-
  Manages an Amazon API Gateway Version 2 authorizer.
---

# Resource: aws_api_gateway_v2_authorizer

Manages an Amazon API Gateway Version 2 authorizer.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_v2_authorizer" "example" {
  api_id           = "${aws_api_gateway_v2_api.example.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.example.invoke_arn}"
  identity_sources = ["route.request.header.Auth"]
  name             = "example-authorizer"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `authorizer_type` - (Required) The authorizer type. Valid values: `REQUEST`.
* `authorizer_uri` - (Required) The authorizer's Uniform Resource Identifier (URI).
For `REQUEST` authorizers this must be a well-formed Lambda function URI, such as the `invoke_arn` attribute of the [`aws_lambda_function`](/docs/providers/aws/r/lambda_function.html) resource.
* `identity_sources` - (Required) The identity sources for which authorization is requested.
For `REQUEST` authorizers the value is a list of one or more mapping expressions of the specified request parameters.
* `name` - (Required) The name of the authorizer.
* `authorizer_credentials_arn` - (Optional) The required credentials as an IAM role for API Gateway to invoke the authorizer.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The authorizer identifier.

## Import

`aws_api_gateway_v2_authorizer` can be imported by using the API identifier and authorizer identifier, e.g.

```
$ terraform import aws_api_gateway_v2_authorizer.example aabbccddee/1122334
```

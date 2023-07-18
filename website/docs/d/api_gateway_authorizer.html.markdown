---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_authorizer"
description: |-
  Provides details about a specific API Gateway Authorizer.
---

# Data Source: aws_api_gateway_authorizer

Provides details about a specific API Gateway Authorizer.

## Example Usage

```terraform
data "aws_api_gateway_authorizer" "example" {
  rest_api_id   = aws_api_gateway_rest_api.example.id
  authorizer_id = data.aws_api_gateway_authorizers.example.ids[0]
}
```

## Argument Reference

The following arguments are required:

* `authorizer_id` - (Required) Authorizer identifier.
* `rest_api_id` - (Required) ID of the associated REST API.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the API Gateway Authorizer.
* `authorizer_credentials` - Credentials required for the authorizer.
* `authorizer_result_ttl_in_seconds` - TTL of cached authorizer results in seconds.
* `authorizer_uri` - Authorizer's Uniform Resource Identifier (URI).
* `identity_source` - Source of the identity in an incoming request.
* `identity_validation_expression` - Validation expression for the incoming identity.
* `name` - Name of the authorizer.
* `provider_arns` - List of the Amazon Cognito user pool ARNs.
* `type` - Type of the authorizer.

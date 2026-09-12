---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_authorizer"
description: |-
  Provides details about an AWS API Gateway V2 Authorizer.
---
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Data Source: aws_apigatewayv2_authorizer

Provides details about a specific Amazon API Gateway Version 2 Authorizer.

## Example Usage

```terraform
data "aws_apigatewayv2_authorizer" "example" {
  api_id        = "aabbccddee"
  authorizer_id = "bbccddeeff"
}
```

## Argument Reference

This data source supports the following arguments:

* `regional` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `api_id` - (Required) API identifier.
* `authorizer_id` - (Required) Authorizer identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `authorizer_credentials_arn` - IAM role used by API Gateway to invoke authorizer.
* `authorizer_payload_format_version` - Version of payload sent to Lambda authorizer.
* `authorizer_result_ttl_in_seconds` - Time to live (TTL) for cached authorizer results.
* `authorizer_type` - Type of the authorizer.
* `authorizer_uri` - A [lambda function uri](https://docs.aws.amazon.com/apigatewayv2/latest/api-reference/apis-apiid-authorizers-authorizerid.html#apis-apiid-authorizers-authorizerid-properties). Applicable for REQUEST authorizers.
* `enable_simple_responses` - Whether a Lambda authorizer response has a simple format.
* `identity_sources` - Source of the identity in an incoming request.
* `jwt_configuration` - Configuration of JWT authorizer.
    * `audience` - List of the intended recipients of the JWT.
    * `issuer` - Base domain of the identity provider issuing JWT.
* `name` - Name of the authorizer.

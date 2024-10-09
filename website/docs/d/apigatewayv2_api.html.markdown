---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_api"
description: |-
  Provides details about a specific Amazon API Gateway Version 2 API.
---

# Data Source: aws_apigatewayv2_api

Provides details about a specific Amazon API Gateway Version 2 API.

## Example Usage

```terraform
data "aws_apigatewayv2_api" "example" {
  api_id = "aabbccddee"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available APIs in the current region.
The given filters must match exactly one API whose data will be exported as attributes.

This data source supports the following arguments:

* `api_id` - (Required) API identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `api_endpoint` - URI of the API, of the form `https://{api-id}.execute-api.{region}.amazonaws.com` for HTTP APIs and `wss://{api-id}.execute-api.{region}.amazonaws.com` for WebSocket APIs.
* `api_key_selection_expression` - An [API key selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-apikey-selection-expressions).
Applicable for WebSocket APIs.
* `arn` - ARN of the API.
* `cors_configuration` - Cross-origin resource sharing (CORS) [configuration](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-cors.html).
Applicable for HTTP APIs.
* `description` - Description of the API.
* `disable_execute_api_endpoint` - Whether clients can invoke the API by using the default `execute-api` endpoint.
* `execution_arn` - ARN prefix to be used in an [`aws_lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn` attribute
or in an [`aws_iam_policy`](/docs/providers/aws/r/iam_policy.html) to authorize access to the [`@connections` API](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-how-to-call-websocket-api-connections.html).
See the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-control-access-iam.html) for details.
* `name` - Name of the API.
* `protocol_type` - API protocol.
* `route_selection_expression` - The [route selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-selection-expressions) for the API.
* `tags` - Map of resource tags.
* `version` - Version identifier for the API.

The `cors_configuration` object supports the following:

* `allow_credentials` - Whether credentials are included in the CORS request.
* `allow_headers` - Set of allowed HTTP headers.
* `allow_methods` - Set of allowed HTTP methods.
* `allow_origins` - Set of allowed origins.
* `expose_headers` - Set of exposed HTTP headers.
* `max_age` - Number of seconds that the browser should cache preflight request results.

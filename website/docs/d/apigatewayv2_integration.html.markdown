---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_integration"
description: |-
  Provides details an specific Amazon API Gateway V2 API Integration.
---

# Data Source: aws_apigatewayv2_integration

Provides details about an AWS API Gateway V2 Integration.

## Example Usage

```terraform
data "aws_apigatewayv2_integration" "example" {
}
```

## Argument Reference

The following arguments are required:

* `api_id` - (Required) Brief description of the optional argument.
* `integration_id` - (Required) Brief description of the required argument.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `api_gateway_managed` - Whether API Gateway manages the integration.
* `connection_id` - ID of VPC link for a private integration.
* `connection_type` - Network connection type of integration.
* `content_handling_strategy` - How to handle response payload content type conversion.
* `credentials_arn` - Credentials required for the integration.
* `description` - Description of integration.
* `integration_method` - Integration's HTTP method.
* `integration_response_selection_expression` - [Integration response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-integration-response-selection-expressions) for the integration.
* `integration_subtype` - AWS service action to invoke. See [reference](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-aws-services-reference.html).
* `integration_type` - Integration type of an integration.
* `integration_uri` - Uri of integration.
* `passthrough_behavior` - Passthrough behavior for incoming requests based on the Content-Type header.
* `payload_format_version` - Version of payload sent to integration.
* `request_parameters` - A key-value map specifying how parameters are mapped and transformed before being sent to the backend integration.
* `request_templates` - A key-value map of [Velocity](https://velocity.apache.org/) templates.
* `response_parameters` - List of objects showing how response is transformed by status code.
    * `mappings` - Key-value mapping response header parameters name to value.
    * `status_code` - Response status code.
* `template_selection_expression` - [Template selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-template-selection-expressions) for the integration.
* `timeout_milliseconds` - Custom timeout in milliseconds for WebSocket APIs.
* `tls_config` - TLS configuration for private integration.
    * `server_name_to_verify` - Server name API gateway uses to verify hostname on integration's certificate.

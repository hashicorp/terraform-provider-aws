---
layout: "aws"
page_title: "AWS: aws_api_gateway_v2_route"
sidebar_current: "docs-aws-resource-api-gateway-v2-route"
description: |-
  Manages an Amazon API Gateway Version 2 route.
---

# Resource: aws_api_gateway_v2_route

Manages an Amazon API Gateway Version 2 route.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_v2_route" "example" {
  api_id    = "${aws_api_gateway_v2_api.example.id}"
  route_key = "$default"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `route_key` - (Required) The route key for the route.
* `api_key_required` - (Optional) Boolean whether an API key is required for the route. Defaults to `false`.
* `authorization_type` - (Optional) The authorization type for the route. Valid values: `NONE`, `AWS_IAM`, `CUSTOM`. Defaults to `NONE`.
* `authorizer_id` - (Optional) The identifier of the [`aws_api_gateway_v2_authorizer`](/docs/providers/aws/r/api_gateway_v2_authorizer.html) resource to be associated with this route, if the authorizationType is `CUSTOM`.
* `model_selection_expression` - (Optional) The [model selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-model-selection-expressions) for the route.
* `operation_name` - (Optional) The operation name for the route.
* `request_models` - (Optional) The request models for the route.
* `route_response_selection_expression` - (Optional) The [route response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-response-selection-expressions) for the route.
* `target` - (Optional) The target for the route.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The route identifier.

## Import

`aws_api_gateway_v2_route` can be imported by using the API identifier and route identifier, e.g.

```
$ terraform import aws_api_gateway_v2_route.example aabbccddee/1122334
```

---
layout: "aws"
page_title: "AWS: aws_api_gateway_v2_api"
sidebar_current: "docs-aws-resource-api-gateway-v2-api"
description: |-
  Manages an Amazon API Gateway Version 2 API.
---

# Resource: aws_api_gateway_v2_api

Manages an Amazon API Gateway Version 2 API.

-> **Note:** Amazon API Gateway Version 2 resources are used for creating and deploying WebSocket APIs. To create and deploy REST APIs, use Amazon API Gateway Version 1 [resources](https://www.terraform.io/docs/providers/aws/r/api_gateway_rest_api.html).

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_v2_api" "example" {
  name                       = "example-websocket-api"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the API.
* `protocol_type` - (Required) The API protocol. Valid values: `WEBSOCKET`.
* `route_selection_expression` - (Required) The [route selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-selection-expressions) for the API.
* `api_key_selection_expression` - (Optional) An [API key selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-apikey-selection-expressions). Valid values: `$context.authorizer.usageIdentifierKey`, `$request.header.x-api-key`. Defaults to `$request.header.x-api-key`.
* `description` - (Optional) The description of the API.
* `version` - (Optional) A version identifier for the API.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The API identifier.
* `api_endpoint` - The URI of the API, of the form `{api-id}.execute-api.{region}.amazonaws.com`.

## Import

`aws_api_gateway_v2_api` can be imported by using the API identifier, e.g.

```
$ terraform import aws_api_gateway_v2_api.example aabbccddee
```

---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_route"
description: |-
  Manages an Amazon API Gateway Version 2 route.
---

# Resource: aws_apigatewayv2_route

Manages an Amazon API Gateway Version 2 route.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html) for [WebSocket](https://docs.aws.amazon.com/apigateway/latest/developerguide/websocket-api-develop-routes.html) and [HTTP](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-routes.html) APIs.

## Example Usage

### Basic

```terraform
resource "aws_apigatewayv2_api" "example" {
  name                       = "example-websocket-api"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

resource "aws_apigatewayv2_route" "example" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "$default"
}
```

### HTTP Proxy Integration

```terraform
resource "aws_apigatewayv2_api" "example" {
  name          = "example-http-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "example" {
  api_id           = aws_apigatewayv2_api.example.id
  integration_type = "HTTP_PROXY"

  integration_method = "ANY"
  integration_uri    = "https://example.com/{proxy}"
}

resource "aws_apigatewayv2_route" "example" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "ANY /example/{proxy+}"

  target = "integrations/${aws_apigatewayv2_integration.example.id}"
}
```

### WebSocket API with Lambda Integration

This example routes the `$connect` and `$disconnect` special routes and a custom `sendmessage` route (selected by the `action` property of the message body) to an existing Lambda function. The function's execution role is granted `execute-api:ManageConnections` so that it can send messages to connected clients using the [`@connections` API](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-how-to-call-websocket-api-connections.html). Clients connect to the stage's `invoke_url`.

```terraform
resource "aws_apigatewayv2_api" "example" {
  name                       = "example-websocket-api"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

resource "aws_apigatewayv2_integration" "example" {
  api_id             = aws_apigatewayv2_api.example.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.example.invoke_arn
}

resource "aws_apigatewayv2_route" "connect" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "$connect"
  target    = "integrations/${aws_apigatewayv2_integration.example.id}"
}

resource "aws_apigatewayv2_route" "disconnect" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "$disconnect"
  target    = "integrations/${aws_apigatewayv2_integration.example.id}"
}

resource "aws_apigatewayv2_route" "sendmessage" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "sendmessage"
  target    = "integrations/${aws_apigatewayv2_integration.example.id}"
}

resource "aws_apigatewayv2_deployment" "example" {
  api_id = aws_apigatewayv2_api.example.id

  triggers = {
    redeployment = sha1(join(",", tolist([
      jsonencode(aws_apigatewayv2_integration.example),
      jsonencode(aws_apigatewayv2_route.connect),
      jsonencode(aws_apigatewayv2_route.disconnect),
      jsonencode(aws_apigatewayv2_route.sendmessage),
    ])))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_apigatewayv2_stage" "example" {
  api_id        = aws_apigatewayv2_api.example.id
  name          = "example"
  deployment_id = aws_apigatewayv2_deployment.example.id
}

resource "aws_lambda_permission" "example" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.example.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.example.execution_arn}/*"
}

data "aws_iam_policy_document" "manage_connections" {
  statement {
    effect    = "Allow"
    actions   = ["execute-api:ManageConnections"]
    resources = ["${aws_apigatewayv2_api.example.execution_arn}/*"]
  }
}

resource "aws_iam_role_policy" "manage_connections" {
  name   = "manage-connections"
  role   = aws_iam_role.example.id
  policy = data.aws_iam_policy_document.manage_connections.json
}
```

## Argument Reference

This resource supports the following arguments:

* `api_id` - (Required) API identifier.
* `api_key_required` - (Optional) Whether an API key is required for the route. Defaults to `false`. Supported only for WebSocket APIs.
* `authorization_scopes` - (Optional) Authorization scopes supported by this route. The scopes are used with a JWT authorizer to authorize the method invocation.
* `authorization_type` - (Optional) Authorization type for the route. For WebSocket APIs, valid values are `NONE` for open access, `AWS_IAM` for using AWS IAM permissions, and `CUSTOM` for using a Lambda authorizer. For HTTP APIs, valid values are `NONE` for open access, `JWT` for using JSON Web Tokens, `AWS_IAM` for using AWS IAM permissions, and `CUSTOM` for using a Lambda authorizer. Defaults to `NONE`.
* `authorizer_id` - (Optional) Identifier of the [`aws_apigatewayv2_authorizer`](apigatewayv2_authorizer.html) resource to be associated with this route.
* `model_selection_expression` - (Optional) [Model selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-model-selection-expressions) for the route. Supported only for WebSocket APIs.
* `operation_name` - (Optional) Operation name for the route. Must be between 1 and 64 characters in length.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `request_models` - (Optional) Request models for the route. Supported only for WebSocket APIs.
* `request_parameter` - (Optional) Request parameters for the route. Supported only for WebSocket APIs. See [`request_parameter` Block](#request_parameter-block) below.
* `route_key` - (Required) Route key for the route. For HTTP APIs, the route key can be either `$default`, or a combination of an HTTP method and resource path, for example, `GET /pets`.
* `route_response_selection_expression` - (Optional) [Route response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-response-selection-expressions) for the route. Supported only for WebSocket APIs.
* `target` - (Optional) Target for the route, of the form `integrations/`*`IntegrationID`*, where *`IntegrationID`* is the identifier of an [`aws_apigatewayv2_integration`](apigatewayv2_integration.html) resource.

### `request_parameter` Block

The `request_parameter` configuration block supports the following arguments:

* `request_parameter_key` - (Required) Request parameter key. This is a [request data mapping parameter](https://docs.aws.amazon.com/apigateway/latest/developerguide/websocket-api-data-mapping.html#websocket-mapping-request-parameters).
* `required` - (Required) Whether the parameter is required.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Route identifier.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_apigatewayv2_route.example
  identity = {
    api_id = "aabbccddee"
    id     = "1122334"
  }
}

resource "aws_apigatewayv2_route" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `api_id` (String) API identifier.
* `id` (String) Route identifier.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_apigatewayv2_route` using `api_id` and `id` (route identifier), delimited by a `/`. For example:

```terraform
import {
  to = aws_apigatewayv2_route.example
  id = "aabbccddee/1122334"
}
```

Using `terraform import`, import `aws_apigatewayv2_route` using `api_id` and `id` (route identifier), delimited by a `/`. For example:

```console
% terraform import aws_apigatewayv2_route.example aabbccddee/1122334
```

-> **Note:** The API Gateway managed route created as part of [*quick_create*](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-basic-concept.html#apigateway-definition-quick-create) cannot be imported.

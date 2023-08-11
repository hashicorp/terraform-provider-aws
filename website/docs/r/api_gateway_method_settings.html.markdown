---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_method_settings"
description: |-
  Manages API Gateway Stage Method Settings
---

# Resource: aws_api_gateway_method_settings

Manages API Gateway Stage Method Settings. For example, CloudWatch logging and metrics.

~> **NOTE:** We recommend using this resource in conjunction with the [`aws_api_gateway_stage` resource](api_gateway_stage.html) instead of a stage managed by the [`aws_api_gateway_deployment` resource](api_gateway_deployment.html) optional `stage_name` argument. Stages managed by the `aws_api_gateway_deployment` resource are recreated on redeployment and this resource will require a second apply to recreate the method settings.

## Example Usage

### End-to-end

An end-to-end example of a REST API configured with OpenAPI can be found in the [`/examples/api-gateway-rest-api-openapi` directory within the GitHub repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/api-gateway-rest-api-openapi).

### Basic Usage

```terraform
resource "aws_api_gateway_rest_api" "example" {
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/path1" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })

  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.example.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}

resource "aws_api_gateway_method_settings" "all" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled = true
    logging_level   = "ERROR"
  }
}

resource "aws_api_gateway_method_settings" "path_specific" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "path1/GET"

  settings {
    metrics_enabled = true
    logging_level   = "INFO"
  }
}
```

### CloudWatch Logging and Tracing

The AWS Console API Gateway Editor displays multiple options for CloudWatch Logs that don't directly map to the options in the AWS API and Terraform. These examples show the `settings` blocks that are equivalent to the options the AWS Console gives for CloudWatch Logs.

#### Off

```terraform
resource "aws_api_gateway_method_settings" "path_specific" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "path1/GET"

  settings {
    logging_level = "OFF"
  }
}
```

#### Errors Only

```terraform
resource "aws_api_gateway_method_settings" "path_specific" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "path1/GET"

  settings {
    logging_level      = "ERROR"
    metrics_enabled    = true
    data_trace_enabled = false
  }
}
```

#### Errors and Info Logs

```terraform
resource "aws_api_gateway_method_settings" "path_specific" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "path1/GET"

  settings {
    logging_level      = "INFO"
    metrics_enabled    = true
    data_trace_enabled = false
  }
}
```

#### Full Request and Response Logs

```terraform
resource "aws_api_gateway_method_settings" "path_specific" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "path1/GET"

  settings {
    logging_level      = "INFO"
    metrics_enabled    = true
    data_trace_enabled = true
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `rest_api_id` - (Required) ID of the REST API
* `stage_name` - (Required) Name of the stage
* `method_path` - (Required) Method path defined as `{resource_path}/{http_method}` for an individual method override, or `*/*` for overriding all methods in the stage. Ensure to trim any leading forward slashes in the path (e.g., `trimprefix(aws_api_gateway_resource.example.path, "/")`).
* `settings` - (Required) Settings block, see below.

### `settings`

* `metrics_enabled` - (Optional) Whether Amazon CloudWatch metrics are enabled for this method.
* `logging_level` - (Optional) Logging level for this method, which effects the log entries pushed to Amazon CloudWatch Logs. The available levels are `OFF`, `ERROR`, and `INFO`.
* `data_trace_enabled` - (Optional) Whether data trace logging is enabled for this method, which effects the log entries pushed to Amazon CloudWatch Logs.
* `throttling_burst_limit` - (Optional) Throttling burst limit. Default: `-1` (throttling disabled).
* `throttling_rate_limit` - (Optional) Throttling rate limit. Default: `-1` (throttling disabled).
* `caching_enabled` - (Optional) Whether responses should be cached and returned for requests. A cache cluster must be enabled on the stage for responses to be cached.
* `cache_ttl_in_seconds` - (Optional) Time to live (TTL), in seconds, for cached responses. The higher the TTL, the longer the response will be cached.
* `cache_data_encrypted` - (Optional) Whether the cached responses are encrypted.
* `require_authorization_for_cache_control` - (Optional) Whether authorization is required for a cache invalidation request.
* `unauthorized_cache_control_header_strategy` - (Optional) How to handle unauthorized requests for cache invalidation. The available values are `FAIL_WITH_403`, `SUCCEED_WITH_RESPONSE_HEADER`, `SUCCEED_WITHOUT_RESPONSE_HEADER`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_api_gateway_method_settings` using `REST-API-ID/STAGE-NAME/METHOD-PATH`. For example:

```terraform
import {
  to = aws_api_gateway_method_settings.example
  id = "12345abcde/example/test/GET"
}
```

Using `terraform import`, import `aws_api_gateway_method_settings` using `REST-API-ID/STAGE-NAME/METHOD-PATH`. For example:

```console
% terraform import aws_api_gateway_method_settings.example 12345abcde/example/test/GET
```

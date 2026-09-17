---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_alias"
description: |-
  Provides details about an AWS Lambda Alias.
---

# Data Source: aws_lambda_alias

Provides details about an AWS Lambda Alias. Use this data source to retrieve information about an existing Lambda function alias for traffic management, deployment strategies, or API integrations.

## Example Usage

### Basic Usage

```terraform
data "aws_lambda_alias" "example" {
  function_name = "my-lambda-function"
  name          = "production"
}

output "alias_arn" {
  value = data.aws_lambda_alias.example.arn
}
```

### API Gateway Integration

```terraform
data "aws_lambda_alias" "api_handler" {
  function_name = "api-handler"
  name          = "live"
}

resource "aws_api_gateway_integration" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  resource_id = aws_api_gateway_resource.example.id
  http_method = aws_api_gateway_method.example.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = data.aws_lambda_alias.api_handler.invoke_arn
}

# Grant API Gateway permission to invoke the alias
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = data.aws_lambda_alias.api_handler.function_name
  principal     = "apigateway.amazonaws.com"
  qualifier     = data.aws_lambda_alias.api_handler.name
  source_arn    = "${aws_api_gateway_rest_api.example.execution_arn}/*/*"
}
```

### Deployment Version Tracking

```terraform
# Get production alias details
data "aws_lambda_alias" "production" {
  function_name = "payment-processor"
  name          = "production"
}

# Get staging alias details
data "aws_lambda_alias" "staging" {
  function_name = "payment-processor"
  name          = "staging"
}

# Compare versions between environments
locals {
  version_drift = data.aws_lambda_alias.production.function_version != data.aws_lambda_alias.staging.function_version
}

output "deployment_status" {
  value = {
    production_version  = data.aws_lambda_alias.production.function_version
    staging_version     = data.aws_lambda_alias.staging.function_version
    version_drift       = local.version_drift
    ready_for_promotion = !local.version_drift
  }
}
```

### EventBridge Rule Target

```terraform
data "aws_lambda_alias" "event_processor" {
  function_name = "event-processor"
  name          = "stable"
}

resource "aws_cloudwatch_event_rule" "example" {
  name        = "capture-events"
  description = "Capture events for processing"

  event_pattern = jsonencode({
    source      = ["myapp.orders"]
    detail-type = ["Order Placed"]
  })
}

resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.example.name
  target_id = "SendToLambda"
  arn       = data.aws_lambda_alias.event_processor.arn
}

resource "aws_lambda_permission" "allow_eventbridge" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = data.aws_lambda_alias.event_processor.function_name
  principal     = "events.amazonaws.com"
  qualifier     = data.aws_lambda_alias.event_processor.name
  source_arn    = aws_cloudwatch_event_rule.example.arn
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name of the aliased Lambda function.
* `name` - (Required) Name of the Lambda alias.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN identifying the Lambda function alias.
* `description` - Description of the alias.
* `function_version` - Lambda function version which the alias uses.
* `invoke_arn` - ARN to be used for invoking Lambda Function from API Gateway - to be used in [`aws_api_gateway_integration`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_integration)'s `uri`.

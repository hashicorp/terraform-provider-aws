---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_usage_plan_assocation"
description: |-
  Provides an API Gateway Usage Plan Association.
---

# Resource: aws_api_gateway_usage_plan_assocation

Provides an API Gateway Usage Plan Association.

## Example Usage

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

resource "aws_api_gateway_stage" "development" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "development"
}

resource "aws_api_gateway_stage" "production" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "production"
}

resource "aws_api_gateway_usage_plan" "example" {
  name         = "my-usage-plan"
  description  = "my description"
  product_code = "MYCODE"

  # Important to prevent overwriting stages linked via aws_api_gateway_usage_plan_assocation resources
	lifecycle {
		ignore_changes = [api_stages]
	}

  quota_settings {
    limit  = 20
    offset = 2
    period = "WEEK"
  }

  throttle_settings {
    burst_limit = 5
    rate_limit  = 10
  }
}

resource "aws_api_gateway_usage_plan_assocation" "example" {
  usage_plan_id = aws_api_gateway_usage_plan.example.id
  api_id = aws_api_gateway_rest_api.example.id
  stage  = aws_api_gateway_stage.production.stage_name
}
```

## Argument Reference

The API Gateway Usage Plan Association argument layout is a structure composed of several sub-resources - these resources are laid out below.

### Top-Level Arguments

* `api_id` (Required) - API Id of the associated API stage in a usage plan.
* `stage` (Required) - API stage name of the associated API stage in a usage plan.
* `throttle` - (Optional) The [throttling limits](#throttle) of the usage plan.

## Import

AWS API Gateway Usage Plan Association can be imported using the `usage_plan_id`, `api_id` and `stage`, e.g.,

```sh
$ terraform import aws_api_gateway_usage_plan.myusageplan <usage_plan_id>/<api_id>/<stage>
```

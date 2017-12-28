---
layout: "aws"
page_title: "AWS: aws_api_gateway_stage"
sidebar_current: "docs-aws-resource-api-gateway-stage"
description: |-
  Provides an API Gateway Stage.
---

# aws_api_gateway_stage

Provides an API Gateway Stage.

## Example Usage

```hcl
resource "aws_api_gateway_stage" "test" {
  stage_name = "prod"
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  deployment_id = "${aws_api_gateway_deployment.test.id}"

  variables = {
    my_var = "normal value"
  }

  canary_settings {
    percent_traffic = 33.33

    stage_variable_overrides = {
      my_var = "overridden value"
      my_new_var = "true"
    }
  }
}

resource "aws_api_gateway_rest_api" "test" {
  name = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = ["aws_api_gateway_integration.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "dev"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  parent_id   = "${aws_api_gateway_rest_api.test.root_resource_id}"
  path_part   = "mytestresource"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  resource_id   = "${aws_api_gateway_resource.test.id}"
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_settings" "s" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "${aws_api_gateway_stage.test.stage_name}"
  method_path = "${aws_api_gateway_resource.test.path_part}/${aws_api_gateway_method.test.http_method}"

  settings {
    metrics_enabled = true
  logging_level = "INFO"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_resource.test.id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  type        = "MOCK"
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The ID of the associated REST API
* `stage_name` - (Required) The name of the stage
* `deployment_id` - (Required) The ID of the deployment that the stage points to
* `cache_cluster_enabled` - (Optional) Specifies whether a cache cluster is enabled for the stage
* `cache_cluster_size` - (Optional) The size of the cache cluster for the stage, if enabled.
	Allowed values include `0.5`, `1.6`, `6.1`, `13.5`, `28.4`, `58.2`, `118` and `237`.
* `canary_settings` - (Optional) A map of settings for a [canary deployment][0]. Fields documented below.
* `client_certificate_id` - (Optional) The identifier of a client certificate for the stage.
* `description` - (Optional) The description of the stage
* `documentation_version` - (Optional) The version of the associated API documentation
* `variables` - (Optional) A map that defines the stage variables

**canary_settings** supports the following:

* `percent_traffic` - (Optional) The percent `0.0` - `100.0` of traffic to divert to the canary deployment.
* `stage_variable_overrides` - (Optional) A map of overridden stage `variables` (including new variables) for the canary deployment.
* `use_stage_cache` - (Optional) Whether the canary deployment uses the stage cache.

[0]: https://docs.aws.amazon.com/apigateway/latest/developerguide/canary-release.html

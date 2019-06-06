---
layout: "aws"
page_title: "AWS: aws_api_gateway_stage"
sidebar_current: "docs-aws_api_gateway_stage"
description: |-
  Get information on an API Gateway Stage
---

# Data Source: aws_api_gateway_stage

Use this data source to get the latest deploy id from a pre-existing stage. For example
to update a stage's settings without overwriting the currently deployed stage.

## Example Usage

```hcl
resource "aws_api_gateway_rest_api" "my_rest_api" {
  name = "my-rest-api"
}

data "aws_api_gateway_stage" "production" {
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  stage_name = "production"
}

resource "aws_api_gateway_stage" "production" {
  stage_name    = "production"
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  deployment_id = "${data.aws_api_gateway_stage.deployment_id}"
  variables     = {
    some_variable = "value"
  }
}
```

## Argument Reference

 * `rest_api_id` - (Required) The ID of the associated REST API
 * `stage_name` - (Required) The name of the stage

## Attributes Reference

 * `deployment_id` - Set to the ID of the currently deployed deployment.

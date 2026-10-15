---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_usage_plan"
description: |-
  Provides an API Gateway Usage Plan.
---

# Resource: aws_api_gateway_usage_plan

Provides an API Gateway Usage Plan.

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

  api_stages {
    api_id = aws_api_gateway_rest_api.example.id
    stage  = aws_api_gateway_stage.development.stage_name
  }

  api_stages {
    api_id = aws_api_gateway_rest_api.example.id
    stage  = aws_api_gateway_stage.production.stage_name
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
```

## Argument Reference

This resource supports the following arguments:

* `api_stages` - (Optional) Associated API stages of the usage plan. See [`api_stages` Block](#api_stages-block) below.
* `description` - (Optional) Description of a usage plan.
* `name` - (Required) Name of the usage plan.
* `product_code` - (Optional) AWS Marketplace product identifier to associate with the usage plan as a SaaS product on AWS Marketplace.
* `quota_settings` - (Optional) Quota settings of the usage plan. See [`quota_settings` Block](#quota_settings-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `throttle_settings` - (Optional) Throttling limits of the usage plan. See [`throttle_settings` Block](#throttle_settings-block) below.

### `api_stages` Block

The following arguments are required:

* `api_id` - (Required) API ID of the associated API stage in a usage plan.
* `stage` - (Required) API stage name of the associated API stage in a usage plan.

The following arguments are optional:

* `throttle` - (Optional) Throttling limits applied to the API stage. See [`throttle` Block](#throttle-block) below.

### `throttle` Block

The following arguments are required:

* `path` - (Required) Method to apply the throttle settings for. Specify the path and method, for example `/test/GET`.

The following arguments are optional:

* `burst_limit` - (Optional) API request burst limit, the maximum rate limit over a time ranging from one to a few seconds, depending upon whether the underlying token bucket is at its full capacity.
* `rate_limit` - (Optional) API request steady-state rate limit.

### `quota_settings` Block

The following arguments are required:

* `limit` - (Required) Maximum number of requests that can be made in a given time period.
* `period` - (Required) Time period in which the limit applies. Valid values are `DAY`, `WEEK`, or `MONTH`.

The following arguments are optional:

* `offset` - (Optional) Number of requests subtracted from the given limit in the initial time period.

### `throttle_settings` Block

* `burst_limit` - (Optional) API request burst limit, the maximum rate limit over a time ranging from one to a few seconds, depending upon whether the underlying token bucket is at its full capacity.
* `rate_limit` - (Optional) API request steady-state rate limit.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `api_stages` - Associated API stages of the usage plan.
* `arn` - ARN of the usage plan.
* `description` - Description of a usage plan.
* `id` - ID of the usage plan.
* `name` - Name of the usage plan.
* `product_code` - AWS Marketplace product identifier associated with the usage plan.
* `quota_settings` - Quota settings of the usage plan.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `throttle_settings` - Throttling limits of the usage plan.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AWS API Gateway Usage Plan using the `id`. For example:

```terraform
import {
  to = aws_api_gateway_usage_plan.myusageplan
  id = "<usage_plan_id>"
}
```

Using `terraform import`, import AWS API Gateway Usage Plan using the `id`. For example:

```console
% terraform import aws_api_gateway_usage_plan.myusageplan <usage_plan_id>
```

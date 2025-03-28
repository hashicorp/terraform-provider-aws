---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_api_put"
description: |-
  Terraform resource for updating an AWS API Gateway REST API.
---
# Resource: aws_api_gateway_rest_api_put

Terraform resource for updating an AWS API Gateway REST API with a new API description.

~> **NOTE:** When importing an OpenAPI specification using the `body` argument, the `info.title` field updates the `name` of the [`aws_api_gateway_rest_api`](/docs/providers/aws/r/api_gateway_rest_api.html). If the imported `title` differs from the configured `name`, Terraform will report the difference.

## Example Usage

### Basic Usage

```terraform
resource "aws_api_gateway_rest_api_put" "example" {
  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "Example API"
      version = "v1"
    }
    schemes = ["https"]
    paths = {
      "/example" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })

  fail_on_warnings = true
  rest_api_id      = aws_api_gateway_rest_api.example.id
}
```

### Multi-stage

This example demonstrates how to use `aws_api_gateway_rest_api_put` to manage blue-green deployments and multi-stage APIs in Amazon API Gateway with Swagger/OpenAPI. The new resource provides native support for a simple and reliable process.

```terraform
resource "aws_api_gateway_rest_api" "example" {
  name = "Example API"
}

resource "aws_api_gateway_rest_api_put" "examplev1" {
  body             = file("v1.yaml")
  fail_on_warnings = true
  rest_api_id      = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(file("v1.yaml"))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_deployment" "examplev1" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = aws_api_gateway_rest_api_put.examplev1.triggers.redeployment
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "examplev1" {
  stage_name    = "v1"
  rest_api_id   = aws_api_gateway_rest_api.example.id
  deployment_id = aws_api_gateway_deployment.examplev1.id
}

resource "aws_api_gateway_rest_api_put" "examplev2" {
  depends_on = [
    aws_api_gateway_stage.examplev1
  ]

  body             = file("v2.yaml")
  fail_on_warnings = true
  rest_api_id      = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(file("v2.yaml"))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_deployment" "examplev2" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = aws_api_gateway_rest_api_put.examplev2.triggers.redeployment
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "examplev2" {
  stage_name    = "v2"
  rest_api_id   = aws_api_gateway_rest_api.example.id
  deployment_id = aws_api_gateway_deployment.examplev2.id
}
```

## Argument Reference

The following arguments are required:

* `body` - (Required) PUT request body containing external API definitions. Currently, only OpenAPI definition JSON/YAML files are supported. The maximum size of the API definition file is 6MB.
* `rest_api_id` - (Required) Identifier of the associated REST API.

The following arguments are optional:

* `fail_on_warnings` - (Optional) Whether to rollback the API update when a warning is encountered. The default value is `false`.
* `parameters` - (Optional) Map of customizations for importing the specification in the `body` argument. For example, to exclude DocumentationParts from an imported API, use `ignore = "documentation"`. Additional documentation, including other parameters such as `basepath`, can be found in the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-import-api.html).
* `triggers` - (Optional) Map of arbitrary keys and values that, when changed, will trigger a redeployment. To force a redeployment without changing these keys/values, use the [`-replace` option](https://developer.hashicorp.com/terraform/cli/commands/plan#replace-address) with `terraform plan` or `terraform apply`.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway REST API Put using the `rest_api_id`. For example:

```terraform
import {
  to = aws_api_gateway_rest_api_put.example
  id = "import-id-12345678"
}
```

Using `terraform import`, import API Gateway REST API Put using the `rest_api_id`. For example:

```console
% terraform import aws_api_gateway_rest_api_put.example import-id-12345678
```

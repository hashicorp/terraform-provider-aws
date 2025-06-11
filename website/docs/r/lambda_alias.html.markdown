---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_alias"
description: |-
  Manages an AWS Lambda Alias.
---

# Resource: aws_lambda_alias

Manages an AWS Lambda Alias. Use this resource to create an alias that points to a specific Lambda function version for traffic management and deployment strategies.

For information about Lambda and how to use it, see [What is AWS Lambda?](http://docs.aws.amazon.com/lambda/latest/dg/welcome.html). For information about function aliases, see [CreateAlias](http://docs.aws.amazon.com/lambda/latest/dg/API_CreateAlias.html) and [AliasRoutingConfiguration](https://docs.aws.amazon.com/lambda/latest/dg/API_AliasRoutingConfiguration.html) in the API docs.

## Example Usage

### Basic Alias

```terraform
resource "aws_lambda_alias" "example" {
  name             = "production"
  description      = "Production environment alias"
  function_name    = aws_lambda_function.example.arn
  function_version = "1"
}
```

### Alias with Traffic Splitting

```terraform
resource "aws_lambda_alias" "example" {
  name             = "staging"
  description      = "Staging environment with traffic splitting"
  function_name    = aws_lambda_function.example.function_name
  function_version = "2"

  routing_config {
    additional_version_weights = {
      "1" = 0.1 # Send 10% of traffic to version 1
      "3" = 0.2 # Send 20% of traffic to version 3
      # Remaining 70% goes to version 2 (the primary version)
    }
  }
}
```

### Blue-Green Deployment Alias

```terraform
# Alias for gradual rollout
resource "aws_lambda_alias" "example" {
  name             = "live"
  description      = "Live traffic with gradual rollout to new version"
  function_name    = aws_lambda_function.example.function_name
  function_version = "5" # Current stable version

  routing_config {
    additional_version_weights = {
      "6" = 0.05 # Send 5% of traffic to new version for testing
    }
  }
}
```

### Development Alias

```terraform
resource "aws_lambda_alias" "example" {
  name             = "dev"
  description      = "Development environment - always points to latest"
  function_name    = aws_lambda_function.example.function_name
  function_version = "$LATEST"
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function.
* `function_version` - (Required) Lambda function version for which you are creating the alias. Pattern: `(\$LATEST|[0-9]+)`.
* `name` - (Required) Name for the alias. Pattern: `(?!^[0-9]+$)([a-zA-Z0-9-_]+)`.

The following arguments are optional:

* `description` - (Optional) Description of the alias.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `routing_config` - (Optional) Lambda alias' route configuration settings. [See below](#routing_config-configuration-block).

### routing_config Configuration Block

* `additional_version_weights` - (Optional) Map that defines the proportion of events that should be sent to different versions of a Lambda function.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN identifying your Lambda function alias.
* `invoke_arn` - ARN to be used for invoking Lambda Function from API Gateway - to be used in [`aws_api_gateway_integration`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_integration)'s `uri`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda Function Aliases using the `function_name/alias`. For example:

```terraform
import {
  to = aws_lambda_alias.example
  id = "example/production"
}
```

For backwards compatibility, the following legacy `terraform import` command is also supported:

```console
% terraform import aws_lambda_alias.example example/production
```

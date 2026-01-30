---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway"
description: |-
  Manages an AWS Bedrock AgentCore Gateway.
---

# Resource: aws_bedrockagentcore_gateway

Manages an AWS Bedrock AgentCore Gateway. With Gateway, developers can convert APIs, Lambda functions, and existing services into Model Context Protocol (MCP)-compatible tools.

## Example Usage

### Gateway with JWT Authorization

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "bedrock-agentcore-gateway-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_bedrockagentcore_gateway" "example" {
  name     = "example-gateway"
  role_arn = aws_iam_role.example.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"
}
```

### Gateway with advanced JWT Authorization and MCP Configuration

```terraform
resource "aws_bedrockagentcore_gateway" "example" {
  name        = "mcp-gateway"
  description = "Gateway for MCP communication"
  role_arn    = aws_iam_role.example.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://auth.example.com/.well-known/openid-configuration"
      allowed_audience = ["app-client", "web-client"]
      allowed_clients  = ["client-123", "client-456"]
    }
  }

  protocol_type = "MCP"
  protocol_configuration {
    mcp {
      instructions       = "Gateway for handling MCP requests"
      search_type        = "HYBRID"
      supported_versions = ["2025-03-26", "2025-06-18"]
    }
  }
}
```

### Gateway with Interceptor Configuration

```terraform
resource "aws_lambda_function" "interceptor" {
  filename      = "interceptor.zip"
  function_name = "gateway-interceptor"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "python3.12"
}

resource "aws_bedrockagentcore_gateway" "example" {
  name     = "gateway-with-interceptor"
  role_arn = aws_iam_role.example.arn

  authorizer_type = "AWS_IAM"
  protocol_type   = "MCP"

  interceptor_configuration {
    interception_points = ["REQUEST", "RESPONSE"]

    interceptor {
      lambda {
        arn = aws_lambda_function.interceptor.arn
      }
    }

    input_configuration {
      pass_request_headers = true
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `authorizer_type` - (Required) Type of authorizer to use. Valid values: `CUSTOM_JWT`, `AWS_IAM`. When set to `CUSTOM_JWT`, `authorizer_configuration` block is required.
* `name` - (Required) Name of the gateway.
* `protocol_type` - (Required) Protocol type for the gateway. Valid values: `MCP`.
* `role_arn` - (Required) ARN of the IAM role that the gateway assumes to access AWS services.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `authorizer_configuration` - (Optional) Configuration for request authorization. Required when `authorizer_type` is set to `CUSTOM_JWT`. See [`authorizer_configuration`](#authorizer_configuration) below.
* `description` - (Optional) Description of the gateway.
* `exception_level` - (Optional) Exception level for the gateway. Valid values: `INFO`, `WARN`, `ERROR`.
* `interceptor_configuration` - (Optional) List of interceptor configurations for the gateway. Minimum of 1, maximum of 2. See [`interceptor_configuration`](#interceptor_configuration) below.
* `kms_key_arn` - (Optional) ARN of the KMS key used to encrypt the gateway data.
* `protocol_configuration` - (Optional) Protocol-specific configuration for the gateway. See [`protocol_configuration`](#protocol_configuration) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authorizer_configuration`

The `authorizer_configuration` block supports the following:

* `custom_jwt_authorizer` - (Required) JWT-based authorization configuration block. See [`custom_jwt_authorizer`](#custom_jwt_authorizer) below.

### `custom_jwt_authorizer`

The `custom_jwt_authorizer` block supports the following:

* `discovery_url` - (Required) URL used to fetch OpenID Connect configuration or authorization server metadata. Must end with `.well-known/openid-configuration`.
* `allowed_audience` - (Optional) Set of allowed audience values for JWT token validation.
* `allowed_clients` - (Optional) Set of allowed client IDs for JWT token validation.

### `interceptor_configuration`

The `interceptor_configuration` block supports the following:

* `interception_points` - (Required) Set of interception points. Valid values: `REQUEST`, `RESPONSE`.
* `interceptor` - (Required) Interceptor infrastructure configuration. See [`interceptor`](#interceptor) below.
* `input_configuration` - (Optional) Input configuration for the interceptor. See [`input_configuration`](#input_configuration) below.

### `interceptor`

The `interceptor` block supports the following:

* `lambda` - (Required) Lambda function configuration for the interceptor. See [`lambda`](#lambda) below.

### `lambda`

The `lambda` block supports the following:

* `arn` - (Required) ARN of the Lambda function to invoke for the interceptor.

### `input_configuration`

The `input_configuration` block supports the following:

* `pass_request_headers` - (Required) Whether to pass request headers to the interceptor.

### `protocol_configuration`

The `protocol_configuration` block supports the following:

* `mcp` - (Optional) Model Context Protocol (MCP) configuration block. See [`mcp`](#mcp) below.

### `mcp`

The `mcp` block supports the following:

* `instructions` - (Optional) Instructions for the MCP protocol configuration.
* `search_type` - (Optional) Search type for MCP. Valid values: `SEMANTIC`.
* `supported_versions` - (Optional) Set of supported MCP protocol versions.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `gateway_arn` - ARN of the Gateway.
* `gateway_id` - Unique identifier of the Gateway.
* `gateway_url` - URL endpoint for the gateway.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `workload_identity_details` - Workload identity details for the gateway. See [`workload_identity_details`](#workload_identity_details) below.

### `workload_identity_details`

The `workload_identity_details` block contains the following:

* `workload_identity_arn` - ARN of the workload identity.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Gateway using the gateway ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway.example
  id = "GATEWAY1234567890"
}
```

Using `terraform import`, import Bedrock AgentCore Gateway using the gateway ID. For example:

```console
% terraform import aws_bedrockagentcore_gateway.example GATEWAY1234567890
```

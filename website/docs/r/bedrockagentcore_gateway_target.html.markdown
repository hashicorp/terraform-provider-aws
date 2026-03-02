---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_target"
description: |-
  Manages an AWS Bedrock AgentCore Gateway Target.
---

# Resource: aws_bedrockagentcore_gateway_target

Manages an AWS Bedrock AgentCore Gateway Target. Gateway targets define the endpoints and configurations that a gateway can invoke, such as Lambda functions or APIs, allowing agents to interact with external services through the Model Context Protocol (MCP).

## Example Usage

### Lambda Target with Gateway IAM Role

```terraform
data "aws_iam_policy_document" "gateway_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "gateway_role" {
  name               = "bedrock-gateway-role"
  assume_role_policy = data.aws_iam_policy_document.gateway_assume.json
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda_role" {
  name               = "example-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_lambda_function" "example" {
  filename      = "example.zip"
  function_name = "example-function"
  role          = aws_iam_role.lambda_role.arn
  handler       = "index.handler"
  runtime       = "nodejs20.x"
}

resource "aws_bedrockagentcore_gateway" "example" {
  name     = "example-gateway"
  role_arn = aws_iam_role.gateway_role.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url = "https://accounts.google.com/.well-known/openid-configuration"
    }
  }
}

resource "aws_bedrockagentcore_gateway_target" "example" {
  name               = "example-target"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  description        = "Lambda function target for processing requests"

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.example.arn

        tool_schema {
          inline_payload {
            name        = "process_request"
            description = "Process incoming requests"

            input_schema {
              type        = "object"
              description = "Request processing schema"

              property {
                name        = "message"
                type        = "string"
                description = "Message to process"
                required    = true
              }

              property {
                name = "options"
                type = "object"

                property {
                  name = "priority"
                  type = "string"
                }

                property {
                  name = "tags"
                  type = "array"

                  items {
                    type = "string"
                  }
                }
              }
            }

            output_schema {
              type = "object"

              property {
                name     = "status"
                type     = "string"
                required = true
              }

              property {
                name = "result"
                type = "string"
              }
            }
          }
        }
      }
    }
  }
}
```

### Target with API Key Authentication

```terraform
resource "aws_bedrockagentcore_gateway_target" "api_key_example" {
  name               = "api-target"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  description        = "External API target with API key authentication"

  credential_provider_configuration {
    api_key {
      provider_arn              = "arn:aws:iam::123456789012:oidc-provider/example.com"
      credential_location       = "HEADER"
      credential_parameter_name = "X-API-Key"
      credential_prefix         = "Bearer"
    }
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.example.arn

        tool_schema {
          inline_payload {
            name        = "api_tool"
            description = "External API integration tool"

            input_schema {
              type        = "string"
              description = "Simple string input for API calls"
            }
          }
        }
      }
    }
  }
}
```

### Target with OAuth Authentication

```terraform
resource "aws_bedrockagentcore_gateway_target" "oauth_example" {
  name               = "oauth-target"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id

  credential_provider_configuration {
    oauth {
      provider_arn       = "arn:aws:iam::123456789012:oidc-provider/oauth.example.com"
      scopes             = ["read", "write"]
      grant_type         = "authorization_code"
      default_return_url = "https://myapp.example.com/callback"

      custom_parameters = {
        "client_type" = "confidential"
      }
    }
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.example.arn

        tool_schema {
          inline_payload {
            name        = "oauth_tool"
            description = "OAuth-authenticated service"

            input_schema {
              type = "array"

              items {
                type = "object"

                property {
                  name     = "id"
                  type     = "string"
                  required = true
                }

                property {
                  name = "value"
                  type = "number"
                }
              }
            }
          }
        }
      }
    }
  }
}
```

### Complex Schema with JSON Serialization

```terraform
resource "aws_bedrockagentcore_gateway_target" "complex_schema" {
  name               = "complex-target"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.example.arn

        tool_schema {
          inline_payload {
            name        = "complex_tool"
            description = "Tool with complex nested schema"

            input_schema {
              type = "object"

              property {
                name = "profile"
                type = "object"

                property {
                  name = "nested_tags"
                  type = "array"
                  items_json = jsonencode({
                    type = "string"
                  })
                }

                property {
                  name = "metadata"
                  type = "object"
                  properties_json = jsonencode({
                    properties = {
                      "created_at" = { type = "string" }
                      "version"    = { type = "number" }
                    }
                    required = ["created_at"]
                  })
                }
              }
            }
          }
        }
      }
    }
  }
}
```

### MCP Server Target with Header Propagation

```terraform
resource "aws_bedrockagentcore_gateway_target" "mcp_with_headers" {
  name               = "mcp-target-with-headers"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  description        = "MCP server target with header propagation"

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://example.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers  = ["x-correlation-id", "x-tenant-id"]
    allowed_response_headers = ["x-rate-limit-remaining"]
    allowed_query_parameters = ["version"]
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the gateway target.
* `gateway_identifier` - (Required) Identifier of the gateway that this target belongs to.
* `target_configuration` - (Required) Configuration for the target endpoint. See [`target_configuration`](#target_configuration) below.

The following arguments are optional:

* `credential_provider_configuration` - (Optional) Configuration for authenticating requests to the target. Required when using `lambda`, `open_api_schema` and `smithy_model` in `mcp` block. If using `mcp_server` in `mcp` block with no authorization, it should not be specified. See [`credential_provider_configuration`](#credential_provider_configuration) below.
* `description` - (Optional) Description of the gateway target.
* `metadata_configuration` - (Optional) Configuration for HTTP header and query parameter propagation between the gateway and target servers. See [`metadata_configuration`](#metadata_configuration) below.
* `region` - (Optional) AWS region where the resource will be created. If not provided, the region from the provider configuration will be used.

### `credential_provider_configuration`

The `credential_provider_configuration` block supports exactly one of the following:

* `gateway_iam_role` - (Optional) Use the gateway's IAM role for authentication. This is an empty configuration block.
* `api_key` - (Optional) API key-based authentication configuration. See [`api_key`](#api_key) below.
* `oauth` - (Optional) OAuth-based authentication configuration. See [`oauth`](#oauth) below.

### `api_key`

The `api_key` block supports the following:

* `provider_arn` - (Required) ARN of the OIDC provider for API key authentication.
* `credential_location` - (Optional) Location where the API key credential is provided. Valid values: `HEADER`, `QUERY_PARAMETER`.
* `credential_parameter_name` - (Optional) Name of the parameter containing the API key credential.
* `credential_prefix` - (Optional) Prefix to add to the API key credential value.

### `oauth`

The `oauth` block supports the following:

* `provider_arn` - (Required) ARN of the Oauth credential provider for OAuth authentication.
* `grant_type` - (Optional) The OAuth grant type. Valid values: `CLIENT_CREDENTIALS` (machine-to-machine authentication), `AUTHORIZATION_CODE` (user-delegated access).
* `default_return_url` - (Optional) The URL where the end user's browser is redirected after obtaining the authorization code. Required when `grant_type` is `AUTHORIZATION_CODE`.
* `scopes` - (Optional) Set of OAuth scopes to request.
* `custom_parameters` - (Optional) Map of custom parameters to include in OAuth requests.

### `metadata_configuration`

The `metadata_configuration` block supports the following:

* `allowed_query_parameters` - (Optional) A set of URL query parameters that are allowed to be propagated from incoming gateway URL to the target. Maximum of 10 parameters.
* `allowed_request_headers` - (Optional) A set of HTTP headers that are allowed to be propagated from incoming client requests to the target. Maximum of 10 headers.
* `allowed_response_headers` - (Optional) A set of HTTP headers that are allowed to be propagated from the target response back to the client. Maximum of 10 headers.

~> **Note:** Header names must contain only alphanumeric characters, hyphens, and underscores. A large number of standard HTTP headers are restricted and cannot be configured for propagation, including authentication, content negotiation, caching, security, CORS, and connection management headers. Headers starting with `X-Amzn-` are prohibited except for `X-Amzn-Bedrock-AgentCore-Runtime-Custom-*` headers. These restrictions are enforced by schema validation. For the full list of restricted headers, see the [AWS documentation](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-headers.html).

### `target_configuration`

The `target_configuration` block supports the following:

* `mcp` - (Optional) Model Context Protocol (MCP) configuration. See [`mcp`](#mcp) below.

### `mcp`

The `mcp` block supports exactly one of the following:

* `lambda` - (Optional) Lambda function target configuration. See [`lambda`](#lambda) below.
* `mcp_server` - (Optional) MCP server target configuration. See [`mcp_server`](#mcp_server) below.
* `open_api_schema` - (Optional) OpenAPI schema-based target configuration. See [`api_schema_configuration`](#api_schema_configuration) below.
* `smithy_model` - (Optional) Smithy model-based target configuration. See [`api_schema_configuration`](#api_schema_configuration) below.

### `lambda`

The `lambda` block supports the following:

* `lambda_arn` - (Required) ARN of the Lambda function to invoke.
* `tool_schema` - (Required) Schema definition for the tool. See [`tool_schema`](#tool_schema) below.

### `tool_schema`

The `tool_schema` block supports exactly one of the following:

* `inline_payload` - (Optional) Inline tool definition. See [`inline_payload`](#inline_payload) below.
* `s3` - (Optional) S3-based tool definition. See [`s3`](#s3) below.

### `inline_payload`

The `inline_payload` block supports the following:

* `name` - (Required) Name of the tool.
* `description` - (Required) Description of what the tool does.
* `input_schema` - (Required) Schema for the tool's input. See [`schema_definition`](#schema_definition) below.
* `output_schema` - (Optional) Schema for the tool's output. See [`schema_definition`](#schema_definition) below.

### `s3`

The `s3` block supports the following:

* `uri` - (Optional) S3 URI where the tool schema is stored.
* `bucket_owner_account_id` - (Optional) Account ID of the S3 bucket owner.

### `mcp_server`

The `mcp_server` block supports the following:

* `endpoint` - (Required) Endpoint for the MCP server target configuration.

### `api_schema_configuration`

The `api_schema_configuration` block supports exactly one of the following:

* `inline_payload` - (Optional) Inline schema payload. See [`inline_payload`](#inline_payload) below.
* `s3` - (Optional) S3-based schema configuration. See [`s3`](#s3) below.

### `inline_payload` (API Schema)

The `inline_payload` block for API schemas supports the following:

* `payload` - (Required) The inline schema payload content.

### `s3` (API Schema)

The `s3` block for API schemas supports the following:

* `uri` - (Optional) S3 URI where the schema is stored.
* `bucket_owner_account_id` - (Optional) Account ID of the S3 bucket owner.

### `schema_definition`

The `schema_definition` block supports the following:

* `type` - (Required) Data type of the schema. Valid values: `string`, `number`, `integer`, `boolean`, `array`, `object`.
* `description` - (Optional) Description of the schema element.
* `items` - (Optional) Schema definition for array items. Can only be used when `type` is `array`. See [`items`](#items) below.
* `property` - (Optional) Set of property definitions for object types. Can only be used when `type` is `object`. See [`property`](#property) below.

### `items`

The `items` block supports the following:

* `type` - (Required) Data type of the array items.
* `description` - (Optional) Description of the array items.
* `items` - (Optional) Nested items definition for arrays of arrays.
* `property` - (Optional) Set of property definitions for arrays of objects. See [`property`](#property) below.

### `property`

The `property` block supports the following:

* `name` - (Required) Name of the property.
* `type` - (Required) Data type of the property.
* `description` - (Optional) Description of the property.
* `required` - (Optional) Whether this property is required. Defaults to `false`.
* `items_json` - (Optional) JSON-encoded schema definition for array items. Used for complex nested structures. Cannot be used with `properties_json`.
* `properties_json` - (Optional) JSON-encoded schema definition for object properties. Used for complex nested structures. Cannot be used with `items_json`.
* `items` - (Optional) Items definition for array properties. See [`items`](#items) above.
* `property` - (Optional) Set of nested property definitions for object properties.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `target_id` - Unique identifier of the gateway target.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Gateway Target using the gateway identifier and target ID separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_target.example
  id = "GATEWAY1234567890,TARGET0987654321"
}
```

Using `terraform import`, import Bedrock AgentCore Gateway Target using the gateway identifier and target ID separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_gateway_target.example GATEWAY1234567890,TARGET0987654321
```

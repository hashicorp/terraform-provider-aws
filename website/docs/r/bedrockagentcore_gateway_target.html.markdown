---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_target"
description: |-
  Manages an AWS Bedrock AgentCore Gateway Target.
---

# Resource: aws_bedrockagentcore_gateway_target

Manages an AWS Bedrock AgentCore Gateway Target. Gateway targets define the endpoints and configurations that a gateway can invoke, such as Lambda functions, APIs, or AgentCore Runtime agents, allowing agents to interact with external services through the Model Context Protocol (MCP) or by routing HTTP traffic directly to a runtime.

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
  runtime       = "nodejs24.x"
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

### Target with IAM SigV4 Authentication (MCP Server)

Use this for `mcp_server` targets pointing at AWS-hosted SigV4-protected endpoints (e.g. another Bedrock AgentCore Runtime). The gateway signs upstream requests using its own IAM role.

```terraform
resource "aws_bedrockagentcore_gateway_target" "sigv4_example" {
  name               = "sigv4-target"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id

  credential_provider_configuration {
    gateway_iam_role {
      service = "bedrock-agentcore"
    }
  }

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://example-runtime.bedrock-agentcore.us-east-1.amazonaws.com/runtimes/example/invocations?qualifier=DEFAULT"
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

### HTTP Target Routing to an AgentCore Runtime

Routes gateway traffic directly to an AgentCore Runtime agent over HTTP, without MCP aggregation. The gateway must not have a `protocol_type` set.

```terraform
resource "aws_bedrockagentcore_agent_runtime" "example" {
  agent_runtime_name = "example-runtime"
  role_arn           = aws_iam_role.runtime_role.arn

  agent_runtime_artifact {
    container_configuration {
      container_uri = "111122223333.dkr.ecr.us-west-2.amazonaws.com/example-runtime:latest"
    }
  }

  network_configuration {
    network_mode = "PUBLIC"
  }
}

resource "aws_bedrockagentcore_gateway_target" "runtime" {
  name               = "runtime-target"
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    http {
      agentcore_runtime {
        arn       = aws_bedrockagentcore_agent_runtime.example.agent_runtime_arn
        qualifier = "DEFAULT"
      }
    }
  }
}
```

### Self-hosted MCP server in a VPC (managed Lattice)

```terraform
resource "aws_bedrockagentcore_gateway_target" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  name               = "my-private-mcp-target"

  target_configuration {
    mcp {
      mcp_server {
        # The MCP server endpoint as seen from inside the VPC.
        endpoint = "https://mcp.internal.example.com/mcp"
      }
    }
  }

  # AgentCore Gateway will provision VPC Lattice ENIs in the specified subnets
  # and route traffic to the MCP server without exposing it to the internet.
  private_endpoint {
    managed_vpc_resource {
      vpc_identifier           = aws_vpc.example.id
      subnet_ids               = aws_subnet.example[*].id
      endpoint_ip_address_type = "IPV4"
      security_group_ids       = [aws_security_group.mcp_lattice.id]
    }
  }
}
```

### Self-hosted MCP server with routing through an internal ALB

Use `routing_domain` when the MCP server has a private TLS certificate. Place an internal ALB with a public ACM certificate in front of the server and set `routing_domain` to the ALB DNS name.

```terraform
resource "aws_bedrockagentcore_gateway_target" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  name               = "my-private-mcp-via-alb"

  target_configuration {
    mcp {
      mcp_server {
        # Must match the domain on the ALB's ACM certificate.
        endpoint = "https://mcp.example.com/mcp"
      }
    }
  }

  private_endpoint {
    managed_vpc_resource {
      vpc_identifier           = aws_vpc.example.id
      subnet_ids               = aws_subnet.example[*].id
      endpoint_ip_address_type = "IPV4"
      # Route through the internal ALB instead of the actual MCP server domain.
      routing_domain = aws_lb.mcp_alb.dns_name
    }
  }
}
```

### Self-managed VPC Lattice resource configuration

```terraform
resource "aws_bedrockagentcore_gateway_target" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  name               = "my-private-mcp-self-managed"

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://mcp.internal.example.com/mcp"
      }
    }
  }

  private_endpoint {
    self_managed_lattice_resource {
      resource_configuration_identifier = aws_vpclattice_resource_configuration.mcp.arn
    }
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
* `private_endpoint` - (Optional) Configuration for private connectivity from AgentCore Gateway to a resource inside your VPC. Traffic is routed through Amazon VPC Lattice and never traverses the public internet. See [`private_endpoint`](#private_endpoint) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `credential_provider_configuration`

The `credential_provider_configuration` block supports exactly one of the following:

* `api_key` - (Optional) API key-based authentication configuration. See [`api_key`](#api_key) below.
* `caller_iam_credentials` - (Optional) Caller IAM credentials-based authentication configuration. See [`caller_iam_credentials`](#caller_iam_credentials) below.
* `gateway_iam_role` - (Optional) Use the gateway's IAM role for authentication. See [`gateway_iam_role`](#gateway_iam_role) below.
* `jwt_passthrough` - (Optional) JWT passthrough-based authentication configuration. This is an empty configuration block.
* `oauth` - (Optional) OAuth-based authentication configuration. See [`oauth`](#oauth) below.

### `api_key`

The `api_key` block supports the following:

* `provider_arn` - (Required) ARN of the OIDC provider for API key authentication.
* `credential_location` - (Optional) Location where the API key credential is provided. Valid values: `HEADER`, `QUERY_PARAMETER`.
* `credential_parameter_name` - (Optional) Name of the parameter containing the API key credential.
* `credential_prefix` - (Optional) Prefix to add to the API key credential value.

### `caller_iam_credentials`

The `caller_iam_credentials` block supports the following:

* `service` - (Required) The service name for the credentials.
* `region` - (Optional) The AWS region for the credentials.

### `oauth`

The `oauth` block supports the following:

* `provider_arn` - (Required) ARN of the Oauth credential provider for OAuth authentication.
* `grant_type` - (Optional) The OAuth grant type. Valid values: `CLIENT_CREDENTIALS` (machine-to-machine authentication), `AUTHORIZATION_CODE` (user-delegated access).
* `default_return_url` - (Optional) The URL where the end user's browser is redirected after obtaining the authorization code. Required when `grant_type` is `AUTHORIZATION_CODE`.
* `scopes` - (Optional) Set of OAuth scopes to request.
* `custom_parameters` - (Optional) Map of custom parameters to include in OAuth requests.

### `gateway_iam_role`

The `gateway_iam_role` block supports the following:

* `region` - (Optional) AWS Region used for SigV4 signing of upstream requests. Defaults to the gateway's Region when omitted. Only meaningful when `service` is set.
* `service` - (Optional) The target AWS service name used for SigV4 signing of upstream requests. Required when calling SigV4-protected endpoints such as another Bedrock AgentCore Runtime (use `bedrock-agentcore`). Omit for non-SigV4 IAM-role-based authentication, in which case the block can be empty (`gateway_iam_role {}`).

### `metadata_configuration`

The `metadata_configuration` block supports the following:

* `allowed_query_parameters` - (Optional) A set of URL query parameters that are allowed to be propagated from incoming gateway URL to the target. Maximum of 10 parameters.
* `allowed_request_headers` - (Optional) A set of HTTP headers that are allowed to be propagated from incoming client requests to the target. Maximum of 10 headers.
* `allowed_response_headers` - (Optional) A set of HTTP headers that are allowed to be propagated from the target response back to the client. Maximum of 10 headers.

~> **Note:** Header names must contain only alphanumeric characters, hyphens, and underscores. A large number of standard HTTP headers are restricted and cannot be configured for propagation, including authentication, content negotiation, caching, security, CORS, and connection management headers. Headers starting with `X-Amzn-` are prohibited except for `X-Amzn-Bedrock-AgentCore-Runtime-Custom-*` headers. These restrictions are enforced by schema validation. For the full list of restricted headers, see the [AWS documentation](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-headers.html).

### `private_endpoint`

The optional `private_endpoint` block configures private connectivity from AgentCore Gateway to a resource inside your VPC. Traffic is routed through [Amazon VPC Lattice](https://docs.aws.amazon.com/vpc-lattice/latest/ug/what-is-vpc-lattice.html) and never traverses the public internet.

Exactly one of `managed_vpc_resource` or `self_managed_lattice_resource` must be specified.

~> **Note:** Gateway targets configured with `private_endpoint` cannot use `NO_AUTH` as the inbound authorizer type on the parent gateway unless an interceptor Lambda is also configured.

* `managed_vpc_resource` - (Optional) AWS creates and manages the VPC Lattice resource gateway and resource configuration on your behalf using a service-linked role. See [`managed_vpc_resource`](#managed_vpc_resource) below.
* `self_managed_lattice_resource` - (Optional) Use an existing VPC Lattice resource configuration that you manage yourself. Useful for cross-account setups or advanced Lattice configurations. See [`self_managed_lattice_resource`](#self_managed_lattice_resource) below.

### `managed_vpc_resource`

The `managed_vpc_resource` block supports the following:

* `vpc_identifier` - (Required) ID of the VPC that contains the private resource.
* `subnet_ids` - (Required) Set of subnet IDs inside the VPC where Lattice ENIs are placed.
* `endpoint_ip_address_type` - (Required) IP address type for the resource configuration endpoint. Valid values: `IPV4`, `IPV6`.
* `security_group_ids` - (Optional) Set of security group IDs (up to 5) to associate with the Lattice resource gateway. Defaults to the VPC default security group.
* `routing_domain` - (Optional) Intermediate domain (e.g. a VPCE or ALB DNS name) to use instead of the actual target domain. Useful when the MCP server uses a private TLS certificate — place an ALB with a public ACM cert in front and set this to the ALB DNS name.
* `tags` - (Optional) Map of tags to apply to the managed Lattice resource gateway.

### `self_managed_lattice_resource`

The `self_managed_lattice_resource` block supports the following:

* `resource_configuration_identifier` - (Required) ARN or ID of the VPC Lattice resource configuration.

### `target_configuration`

The `target_configuration` block supports exactly one of the following:

* `mcp` - (Optional) Model Context Protocol (MCP) configuration. See [`mcp`](#mcp) below.
* `http` - (Optional) HTTP target configuration for routing requests directly to an AgentCore Runtime agent. See [`http`](#http) below.

### `mcp`

The `mcp` block supports exactly one of the following:

* `api_gateway` - (Optional) API Gateway target configuration. See [`api_gateway`](#api_gateway) below.
* `lambda` - (Optional) Lambda function target configuration. See [`lambda`](#lambda) below.
* `mcp_server` - (Optional) MCP server target configuration. See [`mcp_server`](#mcp_server) below.
* `open_api_schema` - (Optional) OpenAPI schema-based target configuration. See [`api_schema_configuration`](#api_schema_configuration) below.
* `smithy_model` - (Optional) Smithy model-based target configuration. See [`api_schema_configuration`](#api_schema_configuration) below.

### `api_gateway`

The `api_gateway` block supports the following:

* `api_gateway_tool_configuration` - (Required) Configuration for API Gateway tools. See [`api_gateway_tool_configuration`](#api_gateway_tool_configuration) below.
* `rest_api_id` - (Required) ID of the API Gateway REST API to invoke.
* `stage` - (Required) Stage name of the REST API to add as a target.

### `api_gateway_tool_configuration`

The `api_gateway_tool_configuration` block supports the following:

* `tool_filter` - (Required) Repeatable block of path and method patterns to expose as tools. See [`tool_filter`](#tool_filter) below.
* `tool_override` - (Required) Repeatable block of explicit tool definitions with optional custom names and descriptions. See [`tool_override`](#tool_override) below.

### `tool_filter`

The `tool_filter` block supports the following:

* `filter_path` - (Required) Resource path to match in the REST API. Supports exact paths (for example, `/pets`) or wildcard paths (for example, `/pets/*` to match all paths under `/pets`). Must match existing paths in the REST API.
* `methods` - (Required) List of HTTP methods to filter for. Valid values: `GET`, `DELETE`, `HEAD`, `OPTIONS`, `PATCH`, `PUT` and `POST`.

### `tool_override`

The `tool_override` block supports the following:

* `description` - (Optional) Description of the tool. Provides information about the purpose and usage of the tool. If not provided, uses the description from the API's OpenAPI specification.
* `method` - (Required) HTTP method to expose for the specified path. Valid values: `GET`, `DELETE`, `HEAD`, `OPTIONS`, `PATCH`, `PUT` and `POST`.
* `name` - (Optional) Name of tool. Identifies the tool in the Model Context Protocol.
* `path` - (Required) Resource path in the REST API (e.g., `/pets`). Must explicitly match an existing path in the REST API.

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
* `listing_mode` - (Optional) Listing mode for the MCP server target. Valid values are `DEFAULT` and `DYNAMIC`. MCP resources for `DEFAULT` targets are cached at the control plane for faster access, while resources for `DYNAMIC` targets are retrieved dynamically when listing tools.

### `http`

The `http` block supports exactly one of the following:

* `agentcore_runtime` - (Optional) AgentCore Runtime target configuration. See [`agentcore_runtime`](#agentcore_runtime) below.
* `passthrough` - (Optional) Passthrough target configuration that forwards requests to an external HTTPS endpoint. See [`passthrough`](#passthrough) below.

~> **Note:** HTTP targets can only be attached to gateways that do not have a `protocol_type` set. They are not supported on MCP-protocol gateways.

### `agentcore_runtime`

The `agentcore_runtime` block supports the following:

* `arn` - (Required) ARN of the AgentCore Runtime agent that the gateway routes requests to.
* `qualifier` - (Optional) Runtime qualifier identifying a specific endpoint version. Defaults to `DEFAULT` when not set.

### `passthrough`

The `passthrough` block supports the following:

* `endpoint` - (Required) HTTPS endpoint that the gateway forwards requests to for this passthrough target. Must start with `https://`.
* `protocol_type` - (Required) Application protocol the passthrough target implements. Valid values: `MCP`, `A2A`, `INFERENCE`, `CUSTOM`.
* `schema` - (Optional) API schema configuration that defines the structure of the passthrough target's API. Supports the same `inline_payload` and `s3` blocks as [`api_schema_configuration`](#api_schema_configuration).
* `stickiness_configuration` - (Optional) Session stickiness configuration routing requests within the same session to the same target. See [`stickiness_configuration`](#stickiness_configuration) below.

### `stickiness_configuration`

The `stickiness_configuration` block supports the following:

* `identifier` - (Required) Expression identifying where to extract the session identifier from the request (for example, `$context.header.x-session-id`).
* `timeout` - (Optional) Session stickiness timeout, in seconds. Valid values range from 1 to 86400.

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

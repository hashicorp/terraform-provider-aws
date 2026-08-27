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

        schema {
          source {
            inline_payload {
              payload = file("${path.module}/runtime-openapi.json")
            }
          }
        }
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

* `gateway_identifier` - (Required) Identifier of the gateway that this target belongs to.
* `name` - (Required) Name of the gateway target.
* `target_configuration` - (Required) Configuration for the target endpoint. See [`target_configuration` Block](#target_configuration-block) below.

The following arguments are optional:

* `credential_provider_configuration` - (Optional) Configuration for authenticating requests to the target. Required when using `lambda`, `open_api_schema` and `smithy_model` in `mcp` block. If using `mcp_server` in `mcp` block with no authorization, it should not be specified. See [`credential_provider_configuration` Block](#credential_provider_configuration-block) below.
* `description` - (Optional) Description of the gateway target.
* `metadata_configuration` - (Optional) Configuration for HTTP header and query parameter propagation between the gateway and target servers. See [`metadata_configuration` Block](#metadata_configuration-block) below.
* `private_endpoint` - (Optional) Configuration for private connectivity from AgentCore Gateway to a resource inside your VPC. Traffic is routed through Amazon VPC Lattice and never traverses the public internet. See [`private_endpoint` Block](#private_endpoint-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `credential_provider_configuration` Block

The `credential_provider_configuration` block supports exactly one of the following:

* `api_key` - (Optional) API key-based authentication configuration. See [`api_key` Block](#api_key-block) below.
* `caller_iam_credentials` - (Optional) Caller IAM credentials-based authentication configuration. See [`caller_iam_credentials` Block](#caller_iam_credentials-block) below.
* `gateway_iam_role` - (Optional) Use the gateway's IAM role for authentication. See [`gateway_iam_role` Block](#gateway_iam_role-block) below.
* `jwt_passthrough` - (Optional) JWT passthrough-based authentication configuration. This is an empty configuration block.
* `oauth` - (Optional) OAuth-based authentication configuration. See [`oauth` Block](#oauth-block) below.

### `api_key` Block

The `api_key` block supports the following:

* `credential_location` - (Optional) Location where the API key credential is provided. Valid values: `HEADER`, `QUERY_PARAMETER`.
* `credential_parameter_name` - (Optional) Name of the parameter containing the API key credential.
* `credential_prefix` - (Optional) Prefix to add to the API key credential value.
* `provider_arn` - (Required) ARN of the OIDC provider for API key authentication.

### `caller_iam_credentials` Block

The `caller_iam_credentials` block supports the following:

* `region` - (Optional) AWS region for the credentials.
* `service` - (Required) Service name for the credentials.

### `oauth` Block

The `oauth` block supports the following:

* `custom_parameters` - (Optional) Map of custom parameters to include in OAuth requests.
* `default_return_url` - (Optional) URL where the end user's browser is redirected after obtaining the authorization code. Required when `grant_type` is `AUTHORIZATION_CODE`.
* `grant_type` - (Optional) OAuth grant type. Valid values: `CLIENT_CREDENTIALS` (machine-to-machine authentication), `AUTHORIZATION_CODE` (user-delegated access).
* `provider_arn` - (Required) ARN of the Oauth credential provider for OAuth authentication.
* `scopes` - (Optional) Set of OAuth scopes to request.

### `gateway_iam_role` Block

The `gateway_iam_role` block supports the following:

* `region` - (Optional) AWS Region used for SigV4 signing of upstream requests. Defaults to the gateway's Region when omitted. Only meaningful when `service` is set.
* `service` - (Optional) Target AWS service name used for SigV4 signing of upstream requests. Required when calling SigV4-protected endpoints such as another Bedrock AgentCore Runtime (use `bedrock-agentcore`). Omit for non-SigV4 IAM-role-based authentication, in which case the block can be empty (`gateway_iam_role {}`).

### `metadata_configuration` Block

The `metadata_configuration` block supports the following:

* `allowed_query_parameters` - (Optional) Set of URL query parameters that are allowed to be propagated from incoming gateway URL to the target. Maximum of 10 parameters.
* `allowed_request_headers` - (Optional) Set of HTTP headers that are allowed to be propagated from incoming client requests to the target. Maximum of 10 headers.
* `allowed_response_headers` - (Optional) Set of HTTP headers that are allowed to be propagated from the target response back to the client. Maximum of 10 headers.

~> **Note:** Header names must contain only alphanumeric characters, hyphens, and underscores. A large number of standard HTTP headers are restricted and cannot be configured for propagation, including authentication, content negotiation, caching, security, CORS, and connection management headers. Headers starting with `X-Amzn-` are prohibited except for `X-Amzn-Bedrock-AgentCore-Runtime-Custom-*` headers. These restrictions are enforced by schema validation. For the full list of restricted headers, see the [AWS documentation](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-headers.html).

### `private_endpoint` Block

The optional `private_endpoint` block configures private connectivity from AgentCore Gateway to a resource inside your VPC. Traffic is routed through [Amazon VPC Lattice](https://docs.aws.amazon.com/vpc-lattice/latest/ug/what-is-vpc-lattice.html) and never traverses the public internet.

Exactly one of `managed_vpc_resource` or `self_managed_lattice_resource` must be specified.

~> **Note:** Gateway targets configured with `private_endpoint` cannot use `NO_AUTH` as the inbound authorizer type on the parent gateway unless an interceptor Lambda is also configured.

* `managed_vpc_resource` - (Optional) AWS creates and manages the VPC Lattice resource gateway and resource configuration on your behalf using a service-linked role. See [`managed_vpc_resource` Block](#managed_vpc_resource-block) below.
* `self_managed_lattice_resource` - (Optional) Use an existing VPC Lattice resource configuration that you manage yourself. Useful for cross-account setups or advanced Lattice configurations. See [`self_managed_lattice_resource` Block](#self_managed_lattice_resource-block) below.

### `managed_vpc_resource` Block

The `managed_vpc_resource` block supports the following:

* `endpoint_ip_address_type` - (Required) IP address type for the resource configuration endpoint. Valid values: `IPV4`, `IPV6`.
* `routing_domain` - (Optional) Intermediate domain (e.g. a VPCE or ALB DNS name) to use instead of the actual target domain. Useful when the MCP server uses a private TLS certificate — place an ALB with a public ACM cert in front and set this to the ALB DNS name.
* `security_group_ids` - (Optional) Set of security group IDs (up to 5) to associate with the Lattice resource gateway. Defaults to the VPC default security group.
* `subnet_ids` - (Required) Set of subnet IDs inside the VPC where Lattice ENIs are placed.
* `tags` - (Optional) Map of tags to apply to the managed Lattice resource gateway.
* `vpc_identifier` - (Required) ID of the VPC that contains the private resource.

### `self_managed_lattice_resource` Block

The `self_managed_lattice_resource` block supports the following:

* `resource_configuration_identifier` - (Required) ARN or ID of the VPC Lattice resource configuration.

### `target_configuration` Block

The `target_configuration` block supports exactly one of the following:

* `http` - (Optional) HTTP target configuration for routing requests directly to an AgentCore Runtime agent. See [`http` Block](#http-block) below.
* `mcp` - (Optional) Model Context Protocol (MCP) configuration. See [`mcp` Block](#mcp-block) below.

### `http` Block

The `http` block supports exactly one of the following:

* `agentcore_runtime` - (Optional) AgentCore Runtime target configuration. See [`agentcore_runtime` Block](#agentcore_runtime-block) below.
* `passthrough` - (Optional) Passthrough target configuration that forwards requests to an external HTTPS endpoint. See [`passthrough` Block](#passthrough-block) below.

~> **Note:** HTTP targets can only be attached to gateways that do not have a `protocol_type` set. They are not supported on MCP-protocol gateways.

### `agentcore_runtime` Block

The `agentcore_runtime` block supports:

* `arn` - (Required) ARN of the AgentCore Runtime agent that the gateway routes requests to.
* `qualifier` - (Optional) Runtime qualifier identifying a specific endpoint version. Defaults to `DEFAULT` when not set.
* `schema` - (Optional) API schema configuration that defines the structure of the runtime target's API. See [`schema` Block](#schema-block) below.

### `schema` Block

The `schema` block supports the following:

* `source` - (Required) Configuration for API schema. See [`api_schema_configuration` Block](#api_schema_configuration-block) below.

### `passthrough` Block

The `passthrough` block supports:

* `endpoint` - (Required) HTTPS endpoint that the gateway forwards requests to for this passthrough target. Must start with `https://`.
* `protocol_type` - (Required) Application protocol the passthrough target implements. Valid values: `MCP`, `A2A`, `INFERENCE`, `CUSTOM`.
* `schema` - (Optional) API schema configuration that defines the structure of the passthrough target's API. Supports the same `inline_payload` and `s3` blocks as [`api_schema_configuration`](#api_schema_configuration).
* `static_query_parameter_conflict_resolution` - (Optional) Controls precedence when a client request supplies a query parameter whose name matches a configured static query parameter. Valid values: `CLIENT_OVERRIDE`, `STATIC_OVERRIDE`.
* `static_query_parameters` - (Optional) Map of static query parameters that the gateway always appends to the outbound URL when forwarding requests to the target.
* `stickiness_configuration` - (Optional) Session stickiness configuration routing requests within the same session to the same target. See [`stickiness_configuration`](#stickiness_configuration) below.

### `stickiness_configuration`

The `stickiness_configuration` block supports the following:

* `composite_identifier` - (Optional) Additional headers to include in session affinity routing.
* `identifier` - (Required) Expression identifying where to extract the session identifier from the request (for example, `$context.header.x-session-id`).
* `timeout` - (Optional) Session stickiness timeout, in seconds. Valid values range from 1 to 86400.

### `mcp` Block

The `mcp` block supports exactly one of the following:

* `api_gateway` - (Optional) API Gateway target configuration. See [`api_gateway` Block](#api_gateway-block) below.
* `connector` - (Optional) Connector integration target configuration. Connectors provide pre-built integrations with AWS services and third-party tools. See [`connector` Block](#connector-block) below.
* `lambda` - (Optional) Lambda function target configuration. See [`lambda` Block](#lambda-block) below.
* `mcp_server` - (Optional) MCP server target configuration. See [`mcp_server` Block](#mcp_server-block) below.
* `open_api_schema` - (Optional) OpenAPI schema-based target configuration. See [`api_schema_configuration` Block](#api_schema_configuration-block) below.
* `smithy_model` - (Optional) Smithy model-based target configuration. See [`api_schema_configuration` Block](#api_schema_configuration-block) below.

### `api_gateway` Block

The `api_gateway` block supports the following:

* `api_gateway_tool_configuration` - (Required) Configuration for API Gateway tools. See [`api_gateway_tool_configuration` Block](#api_gateway_tool_configuration-block) below.
* `rest_api_id` - (Required) ID of the API Gateway REST API to invoke.
* `stage` - (Required) Stage name of the REST API to add as a target.

### `api_gateway_tool_configuration` Block

The `api_gateway_tool_configuration` block supports the following:

* `tool_filter` - (Required) Repeatable block of path and method patterns to expose as tools. See [`tool_filter` Block](#tool_filter-block) below.
* `tool_override` - (Required) Repeatable block of explicit tool definitions with optional custom names and descriptions. See [`tool_override` Block](#tool_override-block) below.

### `tool_filter` Block

The `tool_filter` block supports the following:

* `filter_path` - (Required) Resource path to match in the REST API. Supports exact paths (for example, `/pets`) or wildcard paths (for example, `/pets/*` to match all paths under `/pets`). Must match existing paths in the REST API.
* `methods` - (Required) List of HTTP methods to filter for. Valid values: `GET`, `DELETE`, `HEAD`, `OPTIONS`, `PATCH`, `PUT` and `POST`.

### `tool_override` Block

The `tool_override` block supports the following:

* `description` - (Optional) Description of the tool. Provides information about the purpose and usage of the tool. If not provided, uses the description from the API's OpenAPI specification.
* `method` - (Required) HTTP method to expose for the specified path. Valid values: `GET`, `DELETE`, `HEAD`, `OPTIONS`, `PATCH`, `PUT` and `POST`.
* `name` - (Optional) Name of tool. Identifies the tool in the Model Context Protocol.
* `path` - (Required) Resource path in the REST API (e.g., `/pets`). Must explicitly match an existing path in the REST API.

### `connector` Block

The `connector` block supports the following:

* `configuration` - (Required) Per-tool configurations for the connector. See [`configuration` Block](#configuration-block) below.
* `enabled` - (Optional) List of tool names to enable from this connector. If omitted, all tools provided by the connector are enabled.
* `source` - (Required) Source configuration identifying which connector to use. See [`source` Block](#source-block) below.

### `configuration` Block

The `configuration` block supports the following:

* `description` - (Optional) Agent-facing description override for this tool.
* `name` - (Required) Tool or operation name (for example, `retrieve` or `webSearch`).
* `parameter_override` - (Optional) Parameter overrides to control parameter visibility and descriptions. See [`parameter_override` Block](#parameter_override-block) below.
* `parameter_values` - (Optional) JSON-encoded parameters to set as fixed or default values when provisioning this tool. Free-form JSON whose schema is defined by the connector.

### `parameter_override` Block

The `parameter_overrides` block supports the following:

* `path` - (Required) JSON Pointer path identifying the parameter (for example, `/numberOfResults` or `/filter`).
* `description` - (Optional) Agent-facing description override for this parameter.
* `visible` - (Optional) Whether this parameter is visible to the agent. If not specified, uses the service default.

### `source` Block

The `source` block supports the following:

* `connector_id` - (Required) Identifier for the connector integration (for example, `bedrock-knowledge-bases`).
* `version` - (Optional) Version of the connector to use (for example, `1.2.0`).

### `lambda` Block

The `lambda` block supports the following:

* `lambda_arn` - (Required) ARN of the Lambda function to invoke.
* `tool_schema` - (Required) Schema definition for the tool. See [`tool_schema` Block](#tool_schema-block) below.

### `tool_schema` Block

The `tool_schema` block supports exactly one of the following:

* `inline_payload` - (Optional) Inline tool definition. See [`inline_payload` Block](#inline_payload-block) below.
* `s3` - (Optional) S3-based tool definition. See [`s3` Block](#s3-block) below.

### `inline_payload` Block

The `inline_payload` block supports the following:

* `description` - (Required) Description of what the tool does.
* `input_schema` - (Required) Schema for the tool's input. See [`schema_definition` Block](#schema_definition-block) below.
* `name` - (Required) Name of the tool.
* `output_schema` - (Optional) Schema for the tool's output. See [`schema_definition` Block](#schema_definition-block) below.

### `s3` Block

The `s3` block supports the following:

* `bucket_owner_account_id` - (Optional) Account ID of the S3 bucket owner.
* `uri` - (Optional) S3 URI where the tool schema is stored.

### `mcp_server` Block

The `mcp_server` block supports the following:

* `endpoint` - (Required) Endpoint for the MCP server target configuration.
* `listing_mode` - (Optional) Listing mode for the MCP server target. Valid values are `DEFAULT` and `DYNAMIC`. MCP resources for `DEFAULT` targets are cached at the control plane for faster access, while resources for `DYNAMIC` targets are retrieved dynamically when listing tools.
* `mcp_tool_schema` - (Optional) Tool schema configuration for the MCP server target. Supported only when the credential provider is configured with an authorization code grant type. When set, dynamic tool discovery and synchronization are disabled. See [`mcp_tool_schema` Block](#mcp_tool_schema-block) below.
* `resource_priority` - (Optional) Priority for resolving MCP server targets with shared resource URIs. Lower values take precedence. Defaults to `1000` when not set.

### `mcp_tool_schema` Block

The `mcp_tool_schema` block supports exactly one of the following:

* `inline_payload` - (Optional) Inline tool schema payload. The `inline_payload` block requires a `payload` (string) containing the MCP tool schema definition.
* `s3` - (Optional) S3 location of the tool schema. See [`s3` Block](#s3-block) below.

### `api_schema_configuration` Block

The `api_schema_configuration` block supports exactly one of the following:

* `inline_payload` - (Optional) Inline schema payload. See [`inline_payload` Block](#inline_payload-block) below.
* `s3` - (Optional) S3-based schema configuration. See [`s3` Block](#s3-block) below.

### `inline_payload` Block

The `inline_payload` block supports the following:

* `payload` - (Required) Inline schema payload content.

### `s3` Block

The `s3` block supports the following:

* `bucket_owner_account_id` - (Optional) Account ID of the S3 bucket owner.
* `uri` - (Optional) S3 URI where the schema is stored.

### `schema_definition` Block

The `schema_definition` block supports the following:

* `description` - (Optional) Description of the schema element.
* `items` - (Optional) Schema definition for array items. Can only be used when `type` is `array`. See [`items` Block](#items-block) below.
* `property` - (Optional) Set of property definitions for object types. Can only be used when `type` is `object`. See [`property` Block](#property-block) below.
* `type` - (Required) Data type of the schema. Valid values: `string`, `number`, `integer`, `boolean`, `array`, `object`.

### `items` Block

The `items` block supports the following:

* `description` - (Optional) Description of the array items.
* `items` - (Optional) Nested items definition for arrays of arrays.
* `property` - (Optional) Set of property definitions for arrays of objects. See [`property` Block](#property-block) below.
* `type` - (Required) Data type of the array items.

### `property` Block

The `property` block supports the following:

* `description` - (Optional) Description of the property.
* `required` - (Optional) Whether this property is required. Defaults to `false`.
* `items` - (Optional) Items definition for array properties. See [`items` Block](#items-block) above.
* `items_json` - (Optional) JSON-encoded schema definition for array items. Used for complex nested structures. Cannot be used with `properties_json`.
* `name` - (Required) Name of the property.
* `properties_json` - (Optional) JSON-encoded schema definition for object properties. Used for complex nested structures. Cannot be used with `items_json`.
* `property` - (Optional) Set of nested property definitions for object properties.
* `type` - (Required) Data type of the property.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `target_id` - Unique identifier of the gateway target.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_target.example
  identity = {
    gateway_identifier = "GATEWAY1234567890"
    target_id          = "TARGET0987654321"
  }
}

resource "aws_bedrockagentcore_gateway_target" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `gateway_identifier` (String) Gateway identifier.
* `target_id` (String) Gateway target ID.

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import gateway targets using `gateway_identifier` and `target_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_target.example
  id = "GATEWAY1234567890,TARGET0987654321"
}
```

Using `terraform import`, import gateway targets using `gateway_identifier` and `target_id` separated by a comma (`,`). For example:

```console
% terraform import aws_bedrockagentcore_gateway_target.example GATEWAY1234567890,TARGET0987654321
```

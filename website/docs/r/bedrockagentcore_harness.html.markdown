---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_harness"
description: |-
  Manages an AWS Bedrock AgentCore Harness.
---

# Resource: aws_bedrockagentcore_harness

Manages an AWS Bedrock AgentCore Harness. A Harness is a managed agent loop that wraps model configuration, tools, skills, memory, and compute environment into a single deployable unit.

## Example Usage

### Basic Usage

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
  name               = "bedrock-agentcore-harness-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "example" {
  role = aws_iam_role.example.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["bedrock:InvokeModel", "bedrock:InvokeModelWithResponseStream"]
      Resource = "*"
    }]
  })
}

resource "aws_bedrockagentcore_harness" "example" {
  harness_name       = "example_harness"
  execution_role_arn = aws_iam_role.example.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }
}
```

### With Tools and Truncation

```terraform
resource "aws_bedrockagentcore_harness" "example" {
  harness_name       = "example_with_tools"
  execution_role_arn = aws_iam_role.example.arn

  model {
    bedrock_model_config {
      model_id    = "anthropic.claude-sonnet-4-20250514"
      temperature = 0.7
      top_p       = 0.9
    }
  }

  system_prompt {
    text = "You are a coding assistant."
  }

  allowed_tools   = ["*"]
  max_iterations  = 10
  max_tokens      = 4096
  timeout_seconds = 300

  tool {
    type = "inline_function"
    name = "get_weather"

    config {
      inline_function {
        description = "Get the current weather for a location"
        input_schema = jsonencode({
          type = "object"
          properties = {
            location = {
              type        = "string"
              description = "City name"
            }
          }
          required = ["location"]
        })
      }
    }
  }

  truncation {
    strategy = "sliding_window"

    config {
      sliding_window {
        messages_count = 50
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `harness_name` - (Required, Forces new resource) Name of the harness. Must be 1-40 characters, alphanumeric and underscores only.
* `execution_role_arn` - (Required) ARN of the IAM role that the harness assumes to access AWS services.
* `model` - (Required) Model configuration for the harness. See [`model`](#model) below.

The following arguments are optional:

* `allowed_tools` - (Optional) List of tool names allowed for the harness. Use `["*"]` to allow all tools.
* `authorizer_configuration` - (Optional) Authorization configuration for authenticating requests. See [`authorizer_configuration`](#authorizer_configuration) below.
* `environment` - (Optional) Compute environment configuration. See [`environment`](#environment) below.
* `environment_artifact` - (Optional) Environment artifact configuration. See [`environment_artifact`](#environment_artifact) below.
* `environment_variables` - (Optional, Sensitive) Map of environment variables.
* `max_iterations` - (Optional) Maximum number of iterations the agent loop can perform.
* `max_tokens` - (Optional) Maximum number of tokens in the model response.
* `memory` - (Optional) Memory configuration. See [`memory`](#memory) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skill` - (Optional) Skill configurations. See [`skill`](#skill) below.
* `system_prompt` - (Optional) System prompt blocks for the harness. See [`system_prompt`](#system_prompt) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout_seconds` - (Optional) Timeout in seconds for the harness execution.
* `timeouts` - (Optional) Configuration block for operation timeouts. See [Timeouts](#timeouts) below.
* `tool` - (Optional) Tool configurations. See [`tool`](#tool) below.
* `truncation` - (Optional) Truncation configuration for conversation history. See [`truncation`](#truncation) below.

### `model`

The `model` block supports exactly one of the following:

* `bedrock_model_config` - (Optional) Amazon Bedrock model configuration. See [`bedrock_model_config`](#bedrock_model_config) below.
* `openai_model_config` - (Optional) OpenAI model configuration. See [`openai_model_config`](#openai_model_config) below.
* `gemini_model_config` - (Optional) Gemini model configuration. See [`gemini_model_config`](#gemini_model_config) below.

### `bedrock_model_config`

* `model_id` - (Required) Bedrock model ID (e.g., `anthropic.claude-sonnet-4-20250514`).
* `max_tokens` - (Optional) Maximum number of tokens to generate.
* `temperature` - (Optional) Temperature for sampling. Must be between 0 and 2.
* `top_p` - (Optional) Top-p (nucleus) sampling parameter. Must be between 0 and 1.

### `openai_model_config`

* `model_id` - (Required) OpenAI model ID.
* `api_key_arn` - (Required) ARN of the secret containing the API key.
* `max_tokens` - (Optional) Maximum number of tokens to generate.
* `temperature` - (Optional) Temperature for sampling.
* `top_p` - (Optional) Top-p sampling parameter.

### `gemini_model_config`

* `model_id` - (Required) Gemini model ID.
* `api_key_arn` - (Required) ARN of the secret containing the API key.
* `max_tokens` - (Optional) Maximum number of tokens to generate.
* `temperature` - (Optional) Temperature for sampling.
* `top_p` - (Optional) Top-p sampling parameter.
* `top_k` - (Optional) Top-k sampling parameter.

### `system_prompt`

* `text` - (Required) Text content of the system prompt.

### `tool`

* `type` - (Required) Type of tool. Valid values: `remote_mcp`, `agentcore_browser`, `agentcore_gateway`, `inline_function`, `agentcore_code_interpreter`.
* `name` - (Optional) Name of the tool.
* `config` - (Optional) Tool-specific configuration. See [`tool config`](#tool-config) below.

### Tool Config

The `config` block supports exactly one of the following:

* `remote_mcp` - (Optional) Remote MCP server configuration. See [`remote_mcp`](#remote_mcp) below.
* `agentcore_browser` - (Optional) AgentCore browser configuration. See [`agentcore_browser`](#agentcore_browser) below.
* `agentcore_gateway` - (Optional) AgentCore gateway configuration. See [`agentcore_gateway`](#agentcore_gateway) below.
* `inline_function` - (Optional) Inline function configuration. See [`inline_function`](#inline_function) below.
* `agentcore_code_interpreter` - (Optional) AgentCore code interpreter configuration. See [`agentcore_code_interpreter`](#agentcore_code_interpreter) below.

### `remote_mcp`

* `url` - (Required, Sensitive) URL of the remote MCP server.
* `headers` - (Optional, Sensitive) Map of HTTP headers to include in requests to the MCP server.

### `agentcore_browser`

* `browser_arn` - (Optional) ARN of the AgentCore browser resource.

### `agentcore_gateway`

* `gateway_arn` - (Required) ARN of the AgentCore gateway resource.
* `outbound_auth` - (Optional) Outbound authentication configuration. See [`outbound_auth`](#outbound_auth) below.

### `outbound_auth`

Exactly one of the following must be specified:

* `aws_iam` - (Optional) Set to `true` to use AWS IAM authentication.
* `none` - (Optional) Set to `true` to disable authentication.
* `oauth` - (Optional) OAuth credential provider configuration. See [`oauth`](#oauth) below.

### `oauth`

* `provider_arn` - (Required) ARN of the OAuth credential provider.
* `scopes` - (Required) List of OAuth scopes.
* `custom_parameters` - (Optional) Map of custom parameters.
* `grant_type` - (Optional) OAuth grant type.
* `default_return_url` - (Optional) Default return URL for OAuth flow.

### `agentcore_code_interpreter`

* `code_interpreter_arn` - (Optional) ARN of the AgentCore code interpreter resource.

### `inline_function`

* `description` - (Required) Description of the inline function.
* `input_schema` - (Required, Sensitive) JSON string defining the input schema for the function.

### `skill`

* `path` - (Required) Path to the skill.

### `truncation`

* `strategy` - (Required) Truncation strategy. Valid values: `sliding_window`, `summarization`, `none`.
* `config` - (Optional) Strategy-specific configuration. See [`truncation config`](#truncation-config) below.

### Truncation Config

The `config` block supports exactly one of the following:

* `sliding_window` - (Optional) Sliding window truncation configuration. See [`sliding_window`](#sliding_window) below.
* `summarization` - (Optional) Summarization truncation configuration. See [`summarization`](#summarization) below.

### `sliding_window`

* `messages_count` - (Optional) Number of recent messages to keep in the conversation window.

### `summarization`

* `summary_ratio` - (Optional) Ratio of the conversation to summarize (0 to 1).
* `preserve_recent_messages` - (Optional) Number of recent messages to preserve without summarization.
* `summarization_system_prompt` - (Optional) Custom system prompt for the summarization model.

### `environment`

* `agentcore_runtime_environment` - (Required) AgentCore runtime environment configuration. See [`agentcore_runtime_environment`](#agentcore_runtime_environment) below.

### `agentcore_runtime_environment`

* `lifecycle_configuration` - (Optional) Lifecycle configuration. See [`lifecycle_configuration`](#lifecycle_configuration) below.
* `network_configuration` - (Optional) Network configuration. See [`network_configuration`](#network_configuration) below.
* `filesystem_configuration` - (Optional) Filesystem configurations. See [`filesystem_configuration`](#filesystem_configuration) below.

### `lifecycle_configuration`

* `idle_runtime_session_timeout` - (Optional) Timeout in seconds for idle sessions.
* `max_lifetime` - (Optional) Maximum lifetime of the instance in seconds.

### `network_configuration`

* `network_mode` - (Required) Network mode. Valid values: `PUBLIC`, `VPC`.
* `network_mode_config` - (Optional) VPC configuration. See [`network_mode_config`](#network_mode_config) below.

### `network_mode_config`

* `security_groups` - (Required) Security groups for the VPC.
* `subnets` - (Required) Subnets for the VPC.

### `filesystem_configuration`

* `session_storage` - (Optional) Session storage configuration. See [`session_storage`](#session_storage) below.

### `session_storage`

* `mount_path` - (Required) Mount path for session storage.

### `environment_artifact`

* `container_configuration` - (Required) Container configuration. See [`container_configuration`](#container_configuration) below.

### `container_configuration`

* `container_uri` - (Required) URI of the container image.

### `authorizer_configuration`

* `custom_jwt_authorizer` - (Optional) JWT-based authorization configuration. See [`custom_jwt_authorizer`](#custom_jwt_authorizer) below.

### `custom_jwt_authorizer`

* `discovery_url` - (Required) URL for OpenID Connect configuration discovery.
* `allowed_audience` - (Optional) Set of allowed audience values for JWT validation.
* `allowed_clients` - (Optional) Set of allowed client IDs for JWT validation.
* `allowed_scopes` - (Optional) Set of allowed scopes for JWT validation.

### `memory`

* `agentcore_memory_configuration` - (Required) AgentCore memory configuration. See [`agentcore_memory_configuration`](#agentcore_memory_configuration) below.

### `agentcore_memory_configuration`

* `arn` - (Required) ARN of the AgentCore memory resource.
* `actor_id` - (Optional) Actor ID for memory sessions.
* `messages_count` - (Optional) Number of messages to retrieve from memory.
* `retrieval_config` - (Optional) Map of retrieval configuration parameters.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Harness.
* `harness_id` - Unique identifier of the Harness.
* `status` - Current status of the Harness.
* `created_at` - Timestamp when the Harness was created.
* `updated_at` - Timestamp when the Harness was last updated.
* `failure_reason` - Reason for failure if the Harness is in a failed state.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `agentcore_runtime_environment` Computed Attributes

* `agent_runtime_arn` - ARN of the created agent runtime.
* `agent_runtime_id` - ID of the created agent runtime.
* `agent_runtime_name` - Name of the created agent runtime.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_harness.example
  identity = {
    "harness_id" = "example-Ab12Cd34Ef"
  }
}

resource "aws_bedrockagentcore_harness" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `harness_id` (String) ID of the harness.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Harnesses using `harness_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_harness.example
  id = "example-Ab12Cd34Ef"
}
```

Using `terraform import`, import Bedrock AgentCore Harnesses using `harness_id`. For example:

```console
% terraform import aws_bedrockagentcore_harness.example example-Ab12Cd34Ef
```

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

### With Managed Memory

```terraform
resource "aws_bedrockagentcore_harness" "example" {
  harness_name       = "my_harness"
  execution_role_arn = aws_iam_role.example.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  memory {
    managed_memory_configuration {
      event_expiry_duration = 14
      strategies            = ["SEMANTIC", "SUMMARIZATION"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `harness_name` - (Required, Forces new resource) Name of the harness. Must be 1-40 characters, alphanumeric and underscores only.
* `execution_role_arn` - (Required) ARN of the IAM role that the harness assumes to access AWS services.
* `model` - (Required) Model configuration for the harness. See [`model` Block](#model-block) below.

The following arguments are optional:

* `allowed_tools` - (Optional) List of tool names allowed for the harness. Use `["*"]` to allow all tools.
* `authorizer_configuration` - (Optional) Authorization configuration for authenticating requests. See [`authorizer_configuration` Block](#authorizer_configuration-block) below.
* `environment` - (Optional) Compute environment configuration. See [`environment` Block](#environment-block) below.
* `environment_artifact` - (Optional) Environment artifact configuration. See [`environment_artifact` Block](#environment_artifact-block) below.
* `environment_variables` - (Optional, Sensitive) Map of environment variables.
* `max_iterations` - (Optional) Maximum number of iterations the agent loop can perform.
* `max_tokens` - (Optional) Maximum number of tokens in the model response.
* `memory` - (Optional) Memory configuration. See [`memory` Block](#memory-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skill` - (Optional) Skill configurations. See [`skill` Block](#skill-block) below.
* `system_prompt` - (Optional) System prompt blocks for the harness. See [`system_prompt` Block](#system_prompt-block) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `timeout_seconds` - (Optional) Timeout in seconds for the harness execution.
* `tool` - (Optional) Tool configurations. See [`tool` Block](#tool-block) below.
* `truncation` - (Optional) Truncation configuration for conversation history. See [`truncation` Block](#truncation-block) below.

### `model` Block

The `model` block supports exactly one of the following:

* `bedrock_model_config` - (Optional) Amazon Bedrock model configuration. See [`bedrock_model_config` Block](#bedrock_model_config-block) below.
* `openai_model_config` - (Optional) OpenAI model configuration. See [`openai_model_config` Block](#openai_model_config-block) below.
* `gemini_model_config` - (Optional) Gemini model configuration. See [`gemini_model_config` Block](#gemini_model_config-block) below.

### `bedrock_model_config` Block

* `model_id` - (Required) Bedrock model ID (e.g., `anthropic.claude-sonnet-4-20250514`).
* `max_tokens` - (Optional) Maximum number of tokens to generate.
* `temperature` - (Optional) Temperature for sampling. Must be between 0 and 2.
* `top_p` - (Optional) Top-p (nucleus) sampling parameter. Must be between 0 and 1.

### `openai_model_config` Block

* `model_id` - (Required) OpenAI model ID.
* `api_key_arn` - (Required) ARN of the secret containing the API key.
* `max_tokens` - (Optional) Maximum number of tokens to generate.
* `temperature` - (Optional) Temperature for sampling.
* `top_p` - (Optional) Top-p sampling parameter.

### `gemini_model_config` Block

* `model_id` - (Required) Gemini model ID.
* `api_key_arn` - (Required) ARN of the secret containing the API key.
* `max_tokens` - (Optional) Maximum number of tokens to generate.
* `temperature` - (Optional) Temperature for sampling.
* `top_p` - (Optional) Top-p sampling parameter.
* `top_k` - (Optional) Top-k sampling parameter.

### `system_prompt` Block

* `text` - (Required, Sensitive) Text content of the system prompt.

### `tool` Block

* `type` - (Required) Type of tool. Valid values: `remote_mcp`, `agentcore_browser`, `agentcore_gateway`, `inline_function`, `agentcore_code_interpreter`.
* `name` - (Optional) Name of the tool.
* `config` - (Optional) Tool-specific configuration. See [`tool config`](#tool-config) below.

### Tool Config

The `config` block supports exactly one of the following:

* `remote_mcp` - (Optional) Remote MCP server configuration. See [`remote_mcp` Block](#remote_mcp-block) below.
* `agentcore_browser` - (Optional) AgentCore browser configuration. See [`agentcore_browser` Block](#agentcore_browser-block) below.
* `agentcore_gateway` - (Optional) AgentCore gateway configuration. See [`agentcore_gateway` Block](#agentcore_gateway-block) below.
* `inline_function` - (Optional) Inline function configuration. See [`inline_function` Block](#inline_function-block) below.
* `agentcore_code_interpreter` - (Optional) AgentCore code interpreter configuration. See [`agentcore_code_interpreter` Block](#agentcore_code_interpreter-block) below.

### `remote_mcp` Block

* `url` - (Required, Sensitive) URL of the remote MCP server.
* `headers` - (Optional, Sensitive) Map of HTTP headers to include in requests to the MCP server.

### `agentcore_browser` Block

* `browser_arn` - (Optional) ARN of the AgentCore browser resource.

### `agentcore_gateway` Block

* `gateway_arn` - (Required) ARN of the AgentCore gateway resource.
* `outbound_auth` - (Optional) Outbound authentication configuration. See [`outbound_auth` Block](#outbound_auth-block) below.

### `outbound_auth` Block

Exactly one of the following must be specified:

* `aws_iam` - (Optional) Set to `true` to use AWS IAM authentication.
* `none` - (Optional) Set to `true` to disable authentication.
* `oauth` - (Optional) OAuth credential provider configuration. See [`oauth` Block](#oauth-block) below.

### `oauth` Block

* `provider_arn` - (Required) ARN of the OAuth credential provider.
* `scopes` - (Required) List of OAuth scopes.
* `custom_parameters` - (Optional) Map of custom parameters.
* `grant_type` - (Optional) OAuth grant type.
* `default_return_url` - (Optional) Default return URL for OAuth flow.

### `agentcore_code_interpreter` Block

* `code_interpreter_arn` - (Optional) ARN of the AgentCore code interpreter resource.

### `inline_function` Block

* `description` - (Required) Description of the inline function.
* `input_schema` - (Required, Sensitive) JSON string defining the input schema for the function.

### `skill` Block

* `path` - (Required) Path to the skill.

### `truncation` Block

* `strategy` - (Required) Truncation strategy. Valid values: `sliding_window`, `summarization`, `none`.
* `config` - (Optional) Strategy-specific configuration. See [`truncation config`](#truncation-config) below.

### Truncation Config

The `config` block supports exactly one of the following:

* `sliding_window` - (Optional) Sliding window truncation configuration. See [`sliding_window` Block](#sliding_window-block) below.
* `summarization` - (Optional) Summarization truncation configuration. See [`summarization` Block](#summarization-block) below.

### `sliding_window` Block

* `messages_count` - (Optional) Number of recent messages to keep in the conversation window.

### `summarization` Block

* `summary_ratio` - (Optional) Ratio of the conversation to summarize (0 to 1).
* `preserve_recent_messages` - (Optional) Number of recent messages to preserve without summarization.
* `summarization_system_prompt` - (Optional) Custom system prompt for the summarization model.

### `environment` Block

* `agentcore_runtime_environment` - (Required) AgentCore runtime environment configuration. See [`agentcore_runtime_environment` Block](#agentcore_runtime_environment-block) below.

### `agentcore_runtime_environment` Block

* `lifecycle_configuration` - (Optional) Lifecycle configuration. See [`lifecycle_configuration` Block](#lifecycle_configuration-block) below.
* `network_configuration` - (Optional) Network configuration. See [`network_configuration` Block](#network_configuration-block) below.
* `filesystem_configuration` - (Optional) Filesystem configurations. See [`filesystem_configuration` Block](#filesystem_configuration-block) below.

### `lifecycle_configuration` Block

* `idle_runtime_session_timeout` - (Optional) Timeout in seconds for idle sessions.
* `max_lifetime` - (Optional) Maximum lifetime of the instance in seconds.

### `network_configuration` Block

* `network_mode` - (Required) Network mode. Valid values: `PUBLIC`, `VPC`.
* `network_mode_config` - (Optional) VPC configuration. See [`network_mode_config` Block](#network_mode_config-block) below.

### `network_mode_config` Block

* `security_groups` - (Required) Security groups for the VPC.
* `subnets` - (Required) Subnets for the VPC.
* `require_service_s3_endpoint` - (Optional) Whether to require an S3 endpoint for the service in the VPC.

### `filesystem_configuration` Block

Each `filesystem_configuration` block describes a single filesystem to mount into the agent runtime. The list can contain up to 5 entries. Each block must specify exactly one of `session_storage`, `s3_files_access_point`, or `efs_access_point`.

* `session_storage` - (Optional) Session storage filesystem providing persistent storage across agent runtime session invocations. Exactly one of `session_storage`, `s3_files_access_point`, or `efs_access_point` must be specified. See [`session_storage` Block](#session_storage-block) below.
* `s3_files_access_point` - (Optional) Amazon S3 Files access point to mount as shared file storage. Exactly one of `session_storage`, `s3_files_access_point`, or `efs_access_point` must be specified. See [`s3_files_access_point` Block](#s3_files_access_point-block) below.
* `efs_access_point` - (Optional) Amazon EFS access point to mount as shared file storage. Exactly one of `session_storage`, `s3_files_access_point`, or `efs_access_point` must be specified. See [`efs_access_point` Block](#efs_access_point-block) below.

### `session_storage` Block

The `session_storage` block supports the following:

* `mount_path` - (Required) Mount path for the session storage filesystem inside the agent runtime. Must be under `/mnt` with exactly one subdirectory level (for example, `/mnt/data`).

### `s3_files_access_point` Block

The `s3_files_access_point` block supports the following:

* `access_point_arn` - (Required) ARN of the Amazon S3 Files access point to mount into the agent runtime.
* `mount_path` - (Required) Mount path for the S3 Files access point inside the agent runtime. Must be under `/mnt` with exactly one subdirectory level (for example, `/mnt/data`).

### `efs_access_point` Block

The `efs_access_point` block supports the following:

* `access_point_arn` - (Required) ARN of the Amazon EFS access point to mount into the agent runtime.
* `mount_path` - (Required) Mount path for the EFS access point inside the agent runtime. Must be under `/mnt` with exactly one subdirectory level (for example, `/mnt/data`).

### `environment_artifact` Block

* `container_configuration` - (Required) Container configuration. See [`container_configuration` Block](#container_configuration-block) below.

### `container_configuration` Block

* `container_uri` - (Required) URI of the container image.

### `authorizer_configuration` Block

The `authorizer_configuration` block supports the following:

* `custom_jwt_authorizer` - (Optional) JWT-based authorization configuration block. See [`custom_jwt_authorizer` Block](#custom_jwt_authorizer-block) below.

### `custom_jwt_authorizer` Block

The `custom_jwt_authorizer` block supports the following:

* `discovery_url` - (Required) URL used to fetch OpenID Connect configuration or authorization server metadata. Must end with `.well-known/openid-configuration`.

* `advertised_scope_mapping` - (Optional) Map associating each scope in `allowed_scopes` with the corresponding scope advertised to clients in OAuth protected resource metadata and `WWW-Authenticate` response headers. Scopes without a mapping entry are advertised unchanged.
* `allowed_audience` - (Optional) Set of allowed audience values for JWT token validation.
* `allowed_clients` - (Optional) Set of allowed client IDs for JWT token validation.
* `allowed_scopes` - (Optional) Set of scopes that are allowed to access the token.
* `allowed_workload_configuration` - (Optional) Configuration restricting which workloads may use this authorizer. See [`allowed_workload_configuration` Block](#allowed_workload_configuration-block) below.
* `custom_claim` - (Optional) Repeatable block to define a custom claim validation name, value, and operation. See [`custom_claim` Block](#custom_claim-block) below.
* `private_endpoint` - (Optional) Private endpoint used to reach the authorization server. See [`private_endpoint` Block](#private_endpoint-block) below.
* `private_endpoint_overrides` - (Optional) Overrides for the private endpoints used to reach the authorization server. See [`private_endpoint_overrides` Block](#private_endpoint_overrides-block) below.

### `allowed_workload_configuration` Block

* `hosting_environment` - (Optional) Hosting environments allowed to use the authorizer. Between 1 and 10 entries. See [`hosting_environment` Block](#hosting_environment-block) below.
* `workload_identities` - (Optional) List of workload identity names allowed to use the authorizer. Between 1 and 10 entries.

### `hosting_environment` Block

* `arn` - (Required) ARN of the hosting environment.

### `private_endpoint_overrides` Block

* `domain` - (Required) Domain the override applies to.
* `private_endpoint` - (Required) Private endpoint configuration. See [`private_endpoint` Block](#private_endpoint-block) below.

### `private_endpoint` Block

Exactly one of the following must be specified:

* `managed_vpc_resource` - (Optional) Managed VPC resource configuration. See [`managed_vpc_resource` Block](#managed_vpc_resource-block) below.
* `self_managed_lattice_resource` - (Optional) Self-managed VPC Lattice resource configuration. See [`self_managed_lattice_resource` Block](#self_managed_lattice_resource-block) below.

### `managed_vpc_resource` Block

* `endpoint_ip_address_type` - (Required) IP address type for the endpoint. Valid values are `IPV4` and `IPV6`.
* `subnet_ids` - (Required) IDs of the subnets for the endpoint.
* `vpc_identifier` - (Required) Identifier of the VPC for the endpoint.
* `routing_domain` - (Optional) Routing domain for the endpoint.
* `security_group_ids` - (Optional) IDs of the security groups for the endpoint.
* `tags` - (Optional) Tags to assign to the managed VPC resource.

### `self_managed_lattice_resource` Block

* `resource_configuration_identifier` - (Required) Identifier of the VPC Lattice resource configuration.

### `custom_claim` Block

The `custom_claim` block supports the following:

* `authorizing_claim_match_value` - (Required) Configuration block to define the value or values to match for and the relationship of the match. See [`authorizing_claim_match_value` Block](#authorizing_claim_match_value-block) below.
* `inbound_token_claim_name` - (Required) Name of the custom claim field to check.
* `inbound_token_claim_value_type` - (Required) Data type of the claim value to check for. Valid values are `STRING` and `STRING_ARRAY`.

### `authorizing_claim_match_value` Block

The `authorizing_claim_match_value` block supports the following:

* `claim_match_operator` - (Required) Relationship between the claim field value and the value or values to match for. Valid values are `EQUALS`, `CONTAINS`, and `CONTAINS_ANY`. `EQUALS` can be used only when `inbound_token_claim_value_type` is `STRING`. `CONTAINS` or `CONTAINS_ANY` can be used only when `inbound_token_claim_value_type` is `STRING_ARRAY`.
* `claim_match_value` - (Required) Value or values to match for. See [`claim_match_value` Block](#claim_match_value-block) below.

### `claim_match_value` Block

The `claim_match_value` block supports the following:

* `match_value_string` - (Optional) String value to match for. Must be specified when `claim_match_operator` is `EQUALS` or `CONTAINS`. Exactly one of `match_value_string` or `match_value_string_list` must be specified.
* `match_value_string_list` - (Optional) List of strings to check for a match. Must be specified when `claim_match_operator` is `CONTAINS_ANY`. Exactly one of `match_value_string` or `match_value_string_list` must be specified.

### `memory` Block

The `memory` block supports one of the following:

* `agentcore_memory_configuration` - (Optional) AgentCore memory configuration. Use this to connect to an existing AgentCore memory resource. See [`agentcore_memory_configuration` Block](#agentcore_memory_configuration-block) below.
* `managed_memory_configuration` - (Optional) Managed memory configuration. Creates and manages a memory resource automatically. See [`managed_memory_configuration` Block](#managed_memory_configuration-block) below.

### `agentcore_memory_configuration` Block

* `arn` - (Required) ARN of the AgentCore memory resource.
* `actor_id` - (Optional) Actor ID for memory sessions.
* `messages_count` - (Optional) Number of messages to retrieve from memory.
* `retrieval_config` - (Optional) Retrieval configuration parameters. See [`retrieval_config` Block](#retrieval_config-block) below.

### `retrieval_config` Block

`retrieval_config` supports the following:

* `map_block_key` - (Required) Key for the retrieval configuration map block.
* `relevance_score` - (Optional) Relevance score threshold. Valid value is between `0` and `1`.
* `strategy_id` - (Optional) ID of the memory strategy.
* `top_k` - (Optional) Number of top results to retrieve.

### `managed_memory_configuration` Block

* `encryption_key_arn` - (Optional) ARN of a customer-managed KMS key used to encrypt the memory. Defaults to an AWS-owned key. Cannot be changed after creation.
* `event_expiry_duration` - (Optional, Computed) Event retention in days. Defaults to `30`.
* `strategies` - (Optional, Computed) Set of strategy types to enable. Valid values are `SEMANTIC`, `SUMMARIZATION`, and `USER_PREFERENCE`. Defaults to `["SEMANTIC", "SUMMARIZATION"]`.

In addition, the following attribute is exported:

* `arn` - ARN of the managed memory resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `harness_id` - Unique identifier of the Harness.
* `arn` - ARN of the Harness.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
    harness_id = "example-Ab12Cd34Ef"
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

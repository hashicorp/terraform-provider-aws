---
subcategory: "Agents for Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Agent.
---
# Resource: aws_bedrockagent_agent

Terraform resource for managing an AWS Agents for Amazon Bedrock Agent.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_agent_trust.json
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_"
}

data "aws_iam_policy_document" "example_agent_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }

    condition {
      test     = "ArnLike"
      values   = ["arn:aws:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "example_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-v2",
    ]
  }
}

resource "aws_iam_role_policy" "example" {
  policy = data.aws_iam_policy_document.example_agent_permissions.json
  role   = aws_iam_role.example.id
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_bedrockagent_agent" "test" {
  agent_name                  = "my-agent-name"
  agent_resource_role_arn     = aws_iam_role.example.arn
  idle_session_ttl_in_seconds = 500
  foundation_model            = "anthropic.claude-v2"
}
```

## Argument Reference

The following arguments are required:

* `agent_name` - (Required) Name for the agent.
* `agent_resource_role_arn` - (Required) ARN of the Role for the agent.
* `foundation_model` - (Required) Foundation model for the agent to use.

The following arguments are optional:

* `customer_encryption_key_arn` - (Optional) ARN of customer manager key to use for encryption.
* `description` - (Optional) Description of the agent.
* `idle_session_ttl_in_seconds` - (Optional) TTL in seconds for the agent to idle.
* `instruction` - (Optional) Instructions to tell agent what it should do.
* `prepare_agent` (Optional) Whether or not to prepare the agent after creation or modification. Defaults to `true`.
* `prompt_override_configuration` (Optional) Prompt override configuration.
* `tags` - (Optional) Key-value tags for the place index. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### prompt_override_configuration

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `prompt_configurations` - (Required) List of prompt configurations.

The following arguments are optional:

* `override_lambda` - (Optional) ARN of Lambda to use when parsing the raw foundation model output.

### prompt_configurations

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `base_prompt_template` - (Required) Prompt template to replace default.
* `parser_mode` - (Required) DEFAULT or OVERRIDDEN to control if the `override_lambda` is used.
* `prompt_creation_mode` - (Required) DEFAULT or OVERRIDDEN to control if the default or provided `base_prompt_template` is used,
* `prompt_state` - (Required) ENABLED or DISABLED to allow the agent to carry out the step in `prompt_type`.
* `prompt_type` - (Required) The step this prompt applies to. Valid values are `PRE_PROCESSING`, `ORCHESTRATION`, `POST_PROCESSING`, and `KNOWLEDGE_BASE_RESPONSE_GENERATION`.
* `inference_configuration` - (Required) Configures inference for the agent

### inference_configuration

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `max_length` - (Required) Maximum number of tokens in the response between 0 and 4096.
* `stop_sequences` - (Required) List of stop sequences that cause the model to stop generating the response.
* `temperature` - (Required) Likelihood of model selecting higher-probability options when generating a response.
* `top_k` - (Required) Defines the number of most-likely candidates the model chooses the next token from.
* `top_p` - (Required) Defines the number of most-likely candidates the model chooses the next token from.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_arn` - ARN of the Agent.
* `agent_id` - ID of the Agent.
* `id` - ID of the Agent.
* `agent_version` - Version of the Agent.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent using the `id`. For example:

```terraform
import {
  to = aws_bedrockagent_agent.example
  id = "agent-abcd1234"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent using the `id`. For example:

```console
% terraform import aws_bedrockagent_agent.example agent-abcd1234
```

---
subcategory: "Bedrock Agents"
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
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

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
      values   = ["arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "example_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-v2",
    ]
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_agent_trust.json
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_"
}

resource "aws_iam_role_policy" "example" {
  policy = data.aws_iam_policy_document.example_agent_permissions.json
  role   = aws_iam_role.example.id
}

resource "aws_bedrockagent_agent" "example" {
  agent_name                  = "my-agent-name"
  agent_resource_role_arn     = aws_iam_role.example.arn
  idle_session_ttl_in_seconds = 500
  foundation_model            = "anthropic.claude-v2"
}
```

## Argument Reference

The following arguments are required:

* `agent_name` - (Required) Name of the agent.
* `agent_resource_role_arn` - (Required) ARN of the IAM role with permissions to invoke API operations on the agent.
* `foundation_model` - (Required) Foundation model used for orchestration by the agent.

The following arguments are optional:

* `customer_encryption_key_arn` - (Optional) ARN of the AWS KMS key that encrypts the agent.
* `description` - (Optional) Description of the agent.
* `idle_session_ttl_in_seconds` - (Optional) Number of seconds for which Amazon Bedrock keeps information about a user's conversation with the agent. A user interaction remains active for the amount of time specified. If no conversation occurs during this time, the session expires and Amazon Bedrock deletes any data provided before the timeout.
* `instruction` - (Optional) Instructions that tell the agent what it should do and how it should interact with users.
* `prepare_agent` (Optional) Whether to prepare the agent after creation or modification. Defaults to `true`.
* `prompt_override_configuration` (Optional) Configurations to override prompt templates in different parts of an agent sequence. For more information, see [Advanced prompts](https://docs.aws.amazon.com/bedrock/latest/userguide/advanced-prompts.html). See [`prompt_override_configuration` Block](#prompt_override_configuration-block) for details.
* `skip_resource_in_use_check` - (Optional) Whether the in-use check is skipped when deleting the agent.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `prompt_override_configuration` Block

The `prompt_override_configuration` configuration block supports the following arguments:

* `prompt_configurations` - (Required) Configurations to override a prompt template in one part of an agent sequence. See [`prompt_configurations` Block](#prompt_configurations-block) for details.
* `override_lambda` - (Optional) ARN of the Lambda function to use when parsing the raw foundation model output in parts of the agent sequence. If you specify this field, at least one of the `prompt_configurations` block must contain a `parser_mode` value that is set to `OVERRIDDEN`.

### `prompt_configurations` Block

The `prompt_configurations` configuration block supports the following arguments:

* `base_prompt_template` - (Required) prompt template with which to replace the default prompt template. You can use placeholder variables in the base prompt template to customize the prompt. For more information, see [Prompt template placeholder variables](https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-placeholders.html).
* `inference_configuration` - (Required) Inference parameters to use when the agent invokes a foundation model in the part of the agent sequence defined by the `prompt_type`. For more information, see [Inference parameters for foundation models](https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html). See [`inference_configuration` Block](#inference_configuration-block) for details.
* `parser_mode` - (Required) Whether to override the default parser Lambda function when parsing the raw foundation model output in the part of the agent sequence defined by the `prompt_type`. If you set the argument as `OVERRIDDEN`, the `override_lambda` argument in the [`prompt_override_configuration`](#prompt_override_configuration-block) block must be specified with the ARN of a Lambda function. Valid values: `DEFAULT`, `OVERRIDDEN`.
* `prompt_creation_mode` - (Required) Whether to override the default prompt template for this `prompt_type`. Set this argument to `OVERRIDDEN` to use the prompt that you provide in the `base_prompt_template`. If you leave it as `DEFAULT`, the agent uses a default prompt template. Valid values: `DEFAULT`, `OVERRIDDEN`.
* `prompt_state` - (Required) Whether to allow the agent to carry out the step specified in the `prompt_type`. If you set this argument to `DISABLED`, the agent skips that step. Valid Values: `ENABLED`, `DISABLED`.
* `prompt_type` - (Required) Step in the agent sequence that this prompt configuration applies to. Valid values: `PRE_PROCESSING`, `ORCHESTRATION`, `POST_PROCESSING`, `KNOWLEDGE_BASE_RESPONSE_GENERATION`.

### `inference_configuration` Block

The `inference_configuration` configuration block supports the following arguments:

* `max_length` - (Required) Maximum number of tokens to allow in the generated response.
* `stop_sequences` - (Required) List of stop sequences. A stop sequence is a sequence of characters that causes the model to stop generating the response.
* `temperature` - (Required) Likelihood of the model selecting higher-probability options while generating a response. A lower value makes the model more likely to choose higher-probability options, while a higher value makes the model more likely to choose lower-probability options.
* `top_k` - (Required) Number of top most-likely candidates, between 0 and 500, from which the model chooses the next token in the sequence.
* `top_p` - (Required) Top percentage of the probability distribution of next tokens, between 0 and 1 (denoting 0% and 100%), from which the model chooses the next token in the sequence.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_arn` - ARN of the agent.
* `agent_id` - Unique identifier of the agent.
* `agent_version` - Version of the agent.
* `id` - Unique identifier of the agent.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent using the agent ID. For example:

```terraform
import {
  to = aws_bedrockagent_agent.example
  id = "GGRRAED6JP"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent using the agent ID. For example:

```console
% terraform import aws_bedrockagent_agent.example GGRRAED6JP
```

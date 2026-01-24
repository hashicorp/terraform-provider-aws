---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_prompt"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Prompt.
---
# Resource: aws_bedrockagent_prompt

Terraform resource for managing an AWS Bedrock Agents Prompt.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagent_prompt" "example" {
  name        = "MyPrompt"
  description = "My prompt description."
}
```

### With Variants

```terraform
resource "aws_bedrockagent_prompt" "example" {
  name            = "MakePlaylist"
  description     = "My first prompt."
  default_variant = "Variant1"

  variant {
    name     = "Variant1"
    model_id = "amazon.titan-text-express-v1"

    inference_configuration {
      text {
        temperature = 0.8
      }
    }

    template_type = "TEXT"
    template_configuration {
      text {
        text = "Make me a {{genre}} playlist consisting of the following number of songs: {{number}}."

        input_variable {
          name = "genre"
        }
        input_variable {
          name = "number"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the prompt.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the prompt.
* `default_variant` - (Optional) Name of the default variant for your prompt.
* `customer_encryption_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key that you encrypted the prompt with.
* `variant` - (Optional) A list of objects, each containing details about a variant of the prompt. See [Variant](#variant) for more information.
* `tags` (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Variant

* `name` - (Required) Name of the prompt variant.
* `model_id` - (Optional) Unique identifier of the model or [inference profile](https://docs.aws.amazon.com/bedrock/latest/userguide/cross-region-inference.html) with which to run inference on the prompt. If this is not supplied, then a `gen_ai_resource` must be defined.
* `template_type` - (Required) Type of prompt template to use. Valid values: `CHAT`, `TEXT`.
* `additional_model_request_fields` - (Optional) Contains model-specific inference configurations that aren’t in the inferenceConfiguration field. To see model-specific inference parameters, see [Inference request parameters and response fields for foundation models](https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html).
* `metadata` - (Optional) A list of objects, each containing a key-value pair that defines a metadata tag and value to attach to a prompt variant. See [Metadata](#metadata) for more information.
* `inference_configuration` - (Optional) Contains inference configurations for the prompt variant. See [Inference Configuration](#inference-configuration) for more information.
* `gen_ai_resource` - (Optional) Specifies a generative AI resource with which to use the prompt. If this is not supplied, then a `gen_ai_resource` must be defined. See [Generative AI Resource](#generative-ai-resource) for more information.
* `template_configuration` - (Optional) Contains configurations for the prompt template. See [Template Configuration](#template-configuration) for more information.

### Metadata

* `key` - (Required) Key of a metadata tag for a prompt variant.
* `value` - (Required) Value of a metadata tag for a prompt variant.

### Inference Configuration

* `text` - (Optional) Contains inference configurations for the prompt variant. See [Text Inference Configuration](#text-inference-configuration) for more information.

#### Text Inference Configuration

* `max_tokens` - (Optional) Maximum number of tokens to return in the response.
* `stop_sequences` - (Optional) List of strings that define sequences after which the model will stop generating.
* `temperature` - (Optional) Controls the randomness of the response. Choose a lower value for more predictable outputs and a higher value for more surprising outputs.
* `top_p` - (Optional) Percentage of most-likely candidates that the model considers for the next token.

### Generative AI Resource

* `agent` - (Optional) Specifies an Amazon Bedrock agent with which to use the prompt. See [Agent Configuration](#agent-configuration) for more information.

#### Agent Configuration

* `agent_identifier` - (Required) ARN of the agent with which to use the prompt.

### Template Configuration

* `text` - (Optional) Contains configurations for the text in a message for a prompt. See [Text Template Configuration](#text-template-configuration)
* `chat` - (Optional) Contains configurations to use the prompt in a conversational format. See [Chat Template Configuration](#chat-template-configuration) for more information.

#### Text Template Configuration

* `text` - (Required) The message for the prompt.
* `input_variable` - (Optional) A list of variables in the prompt template. See [Input Variable](#input-variable) for more information.
* `cache_point` - (Optional) A cache checkpoint within a template configuration. See [Cache Point](#cache-point) for more information.

#### Chat Template Configuration

* `input_variable` - (Optional) A list of variables in the prompt template. See [Input Variable](#input-variable) for more information.
* `message` - (Optional) A list of messages in the chat for the prompt. See [Message](#message) for more information.
* `system` - (Optional) A list of system prompts to provide context to the model or to describe how it should behave. See [System](#system) for more information.
* `tool_configuration` - (Optional) Configuration information for the tools that the model can use when generating a response. See [Tool Configuration](#tool-configuration) for more information.

#### Message

* `role` - (Required) The role that the message belongs to.
* `content` - (Required) Contains the content for the message you pass to, or receive from a model. See [Message Content] for more information.

#### Message Content

* `cache_point` - (Optional) Creates a cache checkpoint within a message. See [Cache Point](#cache-point) for more information.
* `text` - (Optional) The text in the message.

#### System

* `cache_point` - (Optional) Creates a cache checkpoint within a tool designation. See [Cache Point](#cache-point) for more information.
* `text` - (Optional) The text in the system prompt.

#### Tool Configuration

* `tool_choice` - (Optional) Defines which tools the model should request when invoked. See [Tool Choice](#tool-choice) for more information.
* `tool` - (Optional) A list of tools to pass to a model. See [Tool](#tool) for more information.

#### Tool Choice

* `any` - (Optional) Defines tools, at least one of which must be requested by the model. No text is generated but the results of tool use are sent back to the model to help generate a response. This object has no fields.
* `auto` - (Optional) Defines tools. The model automatically decides whether to call a tool or to generate text instead. This object has no fields.
* `tool` - (Optional) Defines a specific tool that the model must request. No text is generated but the results of tool use are sent back to the model to help generate a response. See [Named Tool](#named-tool) for more information.

#### Named Tool

* `name` - (Required) The name of the tool.

#### Tool

* `cache_point` - (Optional) Creates a cache checkpoint within a tool designation. See [Cache Point](#cache-point) for more information.
* `tool_spec` - (Optional) The specification for the tool. See [Tool Specification](#tool-specification) for more information.

#### Tool Specification

* `name` - (Required) The name of the tool.
* `description` - (Optional) The description of the tool.
* `input_schema` - (Optional) The input schema of the tool. See [Tool Input Schema](#tool-input-schema) for more information.

#### Tool Input Schema

* `json` - (Optional) A JSON object defining the input schema for the tool.

#### Input Variable

* `name` - (Required) The name of the variable.

#### Cache Point

* `type` - (Required) Indicates that the CachePointBlock is of the default type. Valid values: `default`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the prompt.
* `id` - Unique identifier of the prompt.
* `version` - Version of the prompt. When you create a prompt, the version created is the `DRAFT` version.
* `created_at` - Time at which the prompt was created.
* `updated_at` -  Time at which the prompt was last updated.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Agents Prompt using the `id`. For example:

```terraform
import {
  to = aws_bedrockagent_prompt.example
  id = "1A2BC3DEFG"
}
```

Using `terraform import`, import Bedrock Agents Prompt using the `id`. For example:

```console
% terraform import aws_bedrockagent_prompt.example 1A2BC3DEFG
```

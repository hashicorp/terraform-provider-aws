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

* `customer_encryption_key_arn` - (Optional) ARN of the KMS key that you encrypted the prompt with.
* `default_variant` - (Optional) Name of the default variant for your prompt.
* `description` - (Optional) Description of the prompt.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `variant` - (Optional) List of objects, each containing details about a variant of the prompt. See [`variant` Block](#variant-block) for more information.

### `variant` Block

* `additional_model_request_fields` - (Optional) Model-specific inference configurations that aren’t in the inferenceConfiguration field. To see model-specific inference parameters, see [Inference request parameters and response fields for foundation models](https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters.html).
* `gen_ai_resource` - (Optional) Generative AI resource with which to use the prompt. If this is not supplied, then a `model_id` must be defined. See [`gen_ai_resource` Block](#gen_ai_resource-block) for more information.
* `inference_configuration` - (Optional) Inference configurations for the prompt variant. See [`inference_configuration` Block](#inference_configuration-block) for more information.
* `metadata` - (Optional) List of objects, each containing a key-value pair that defines a metadata tag and value to attach to a prompt variant. See [`metadata` Block](#metadata-block) for more information.
* `model_id` - (Optional) Unique identifier of the model or [inference profile](https://docs.aws.amazon.com/bedrock/latest/userguide/cross-region-inference.html) with which to run inference on the prompt. If this is not supplied, then a `gen_ai_resource` must be defined.
* `name` - (Required) Name of the prompt variant.
* `template_configuration` - (Optional) Configurations for the prompt template. See [`template_configuration` Block](#template_configuration-block) for more information.
* `template_type` - (Required) Type of prompt template to use. Valid values: `CHAT`, `TEXT`.

### `gen_ai_resource` Block

* `agent` - (Optional) Amazon Bedrock agent with which to use the prompt. See [`agent` Block](#agent-block) for more information.

### `agent` Block

* `agent_identifier` - (Required) ARN of the agent with which to use the prompt.

### `inference_configuration` Block

* `text` - (Optional) Inference configurations for the prompt variant. See [`variant.inference_configuration.text` Block](#variantinference_configurationtext-block) for more information.

### `variant.inference_configuration.text` Block

* `max_tokens` - (Optional) Maximum number of tokens to return in the response.
* `stop_sequences` - (Optional) List of strings that define sequences after which the model will stop generating.
* `temperature` - (Optional) Controls the randomness of the response. Choose a lower value for more predictable outputs and a higher value for more surprising outputs.
* `top_p` - (Optional) Percentage of most-likely candidates that the model considers for the next token.

### `metadata` Block

* `key` - (Required) Key of a metadata tag for a prompt variant.
* `value` - (Required) Value of a metadata tag for a prompt variant.

### `template_configuration` Block

* `chat` - (Optional) Configurations to use the prompt in a conversational format. See [`chat` Block](#chat-block) for more information.
* `text` - (Optional) Configurations for the text in a message for a prompt. See [`variant.template_configuration.text` Block](#varianttemplate_configurationtext-block) for more information.

### `chat` Block

* `input_variable` - (Optional) List of variables in the prompt template. See [`input_variable` Block](#input_variable-block) for more information.
* `message` - (Optional) List of messages in the chat for the prompt. See [`message` Block](#message-block) for more information.
* `system` - (Optional) List of system prompts to provide context to the model or to describe how it should behave. See [`system` Block](#system-block) for more information.
* `tool_configuration` - (Optional) Configuration information for the tools that the model can use when generating a response. See [`tool_configuration` Block](#tool_configuration-block) for more information.

### `message` Block

* `content` - (Required) Content for the message you pass to, or receive from a model. See [`content` Block](#content-block) for more information.
* `role` - (Required) Role that the message belongs to.

### `content` Block

* `cache_point` - (Optional) Cache checkpoint within a message. See [`cache_point` Block](#cache_point-block) for more information.
* `text` - (Optional) Text in the message.

### `system` Block

* `cache_point` - (Optional) Cache checkpoint within the system prompt. See [`cache_point` Block](#cache_point-block) for more information.
* `text` - (Optional) Text in the system prompt.

### `tool_configuration` Block

* `tool` - (Optional) List of tools to pass to a model. See [`variant.template_configuration.chat.tool_configuration.tool` Block](#varianttemplate_configurationchattool_configurationtool-block) for more information.
* `tool_choice` - (Optional) Configuration for which tools the model should request when invoked. See [`tool_choice` Block](#tool_choice-block) for more information.

### `variant.template_configuration.chat.tool_configuration.tool` Block

* `cache_point` - (Optional) Cache checkpoint within a tool designation. See [`cache_point` Block](#cache_point-block) for more information.
* `tool_spec` - (Optional) Specification for the tool. See [`tool_spec` Block](#tool_spec-block) for more information.

### `tool_spec` Block

* `description` - (Optional) Description of the tool.
* `input_schema` - (Optional) Input schema of the tool. See [`input_schema` Block](#input_schema-block) for more information.
* `name` - (Required) Name of the tool.

### `input_schema` Block

* `json` - (Optional) JSON object defining the input schema for the tool.

### `tool_choice` Block

* `any` - (Optional) Tools, at least one of which must be requested by the model. No text is generated but the results of tool use are sent back to the model to help generate a response. This object has no fields.
* `auto` - (Optional) Tools from which the model automatically decides whether to call a tool or to generate text instead. This object has no fields.
* `tool` - (Optional) Specific tool that the model must request. No text is generated but the results of tool use are sent back to the model to help generate a response. See [`variant.template_configuration.chat.tool_configuration.tool_choice.tool` Block](#varianttemplate_configurationchattool_configurationtool_choicetool-block) for more information.

### `variant.template_configuration.chat.tool_configuration.tool_choice.tool` Block

* `name` - (Required) Name of the tool.

### `variant.template_configuration.text` Block

* `cache_point` - (Optional) Cache checkpoint within a template configuration. See [`cache_point` Block](#cache_point-block) for more information.
* `input_variable` - (Optional) List of variables in the prompt template. See [`input_variable` Block](#input_variable-block) for more information.
* `text` - (Required) Message for the prompt.

### `input_variable` Block

* `name` - (Required) Name of the variable.

### `cache_point` Block

* `type` - (Required) Cache point type. Valid values: `default`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the prompt.
* `created_at` - Time at which the prompt was created.
* `id` - Unique identifier of the prompt.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time at which the prompt was last updated.
* `version` - Version of the prompt. When you create a prompt, the version created is the `DRAFT` version.

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

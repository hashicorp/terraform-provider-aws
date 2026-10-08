---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_flow"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Flow.
---

# Resource: aws_bedrockagent_flow

Terraform resource for managing an AWS Bedrock Agents Flow.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagent_flow" "example" {
  name               = "example-flow"
  execution_role_arn = aws_iam_role.example.arn
}
```

### Default definition

```terraform
resource "aws_bedrockagent_flow" "example" {
  name               = "example"
  execution_role_arn = aws_iam_role.example.arn

  definition {
    connection {
      name   = "FlowInputNodeFlowInputNode0ToPrompt_1PromptsNode0"
      source = "FlowInputNode"
      target = "Prompt_1"
      type   = "Data"

      configuration {
        data {
          source_output = "document"
          target_input  = "topic"
        }
      }
    }
    connection {
      name   = "Prompt_1PromptsNode0ToFlowOutputNodeFlowOutputNode0"
      source = "Prompt_1"
      target = "FlowOutputNode"
      type   = "Data"

      configuration {
        data {
          source_output = "modelCompletion"
          target_input  = "document"
        }
      }
    }
    node {
      name = "FlowInputNode"
      type = "Input"

      configuration {
        input {}
      }

      output {
        name = "document"
        type = "String"
      }
    }
    node {
      name = "Prompt_1"
      type = "Prompt"

      configuration {
        prompt {
          source_configuration {
            inline {
              model_id      = "amazon.titan-text-express-v1"
              template_type = "TEXT"

              inference_configuration {
                text {
                  max_tokens     = 2048
                  stop_sequences = ["User:"]
                  temperature    = 0
                  top_p          = 0.8999999761581421
                }
              }

              template_configuration {
                text {
                  text = "Write a paragraph about {{topic}}."

                  input_variable {
                    name = "topic"
                  }
                }
              }
            }
          }
        }
      }

      input {
        expression = "$.data"
        name       = "topic"
        type       = "String"
      }

      output {
        name = "modelCompletion"
        type = "String"
      }
    }
    node {
      name = "FlowOutputNode"
      type = "Output"

      configuration {
        output {}
      }

      input {
        expression = "$.data"
        name       = "document"
        type       = "String"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `execution_role_arn` - (Required) ARN of the service role with permissions to create and manage a flow. For more information, see [Create a service role for flows in Amazon Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-permissions.html) in the Amazon Bedrock User Guide.
* `name` - (Required) Name for the flow.

The following arguments are optional:

* `customer_encryption_key_arn` - (Optional) ARN of the KMS key to encrypt the flow.
* `definition` - (Optional) Nodes and connections between nodes in the flow. See [`definition` Block](#definition-block) for details.
* `description` - (Optional) Description for the flow.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `definition` Block

The `definition` configuration block supports the following arguments:

* `connection` - (Optional) List of connection definitions in the flow. See [`definition.connection` Block](#definitionconnection-block) for details.
* `node` - (Optional) List of node definitions in the flow. See [`definition.node` Block](#definitionnode-block) for details.

### `definition.connection` Block

The `definition.connection` configuration block supports the following arguments:

* `configuration` - (Required) Configuration of the connection. See [`definition.connection.configuration` Block](#definitionconnectionconfiguration-block) for details.
* `name` - (Required) Name for the connection that you can reference.
* `source` - (Required) Node that the connection starts at.
* `target` - (Required) Node that the connection ends at.
* `type` - (Required) Whether the source node that the connection begins from is a condition node (`Conditional`) or not (`Data`).

### `definition.connection.configuration` Block

The `definition.connection.configuration` configuration block supports the following arguments:

* `conditional` - (Optional) Configuration of a connection originating from a Condition node. See [`definition.connection.configuration.conditional` Block](#definitionconnectionconfigurationconditional-block) for details.
* `data` - (Optional) Configuration of a connection originating from a node that isn't a Condition node. See [`definition.connection.configuration.data` Block](#definitionconnectionconfigurationdata-block) for details.

### `definition.connection.configuration.conditional` Block

The `definition.connection.configuration.conditional` configuration block supports the following arguments:

* `condition` - (Required) Condition that triggers this connection. For more information about how to write conditions, see the Condition node type in the [Node types](https://docs.aws.amazon.com/bedrock/latest/userguide/node-types.html) topic in the Amazon Bedrock User Guide.

### `definition.connection.configuration.data` Block

The `definition.connection.configuration.data` configuration block supports the following arguments:

* `source_output` - (Required) Name of the output in the source node that the connection begins from.
* `target_input` - (Required) Name of the input in the target node that the connection ends at.

### `definition.node` Block

The `definition.node` configuration block supports the following arguments:

* `configuration` - (Required) Configurations for the node. See [`definition.node.configuration` Block](#definitionnodeconfiguration-block) for details.
* `input` - (Optional) Information about an input into the node. See [`definition.node.input` Block](#definitionnodeinput-block) for details.
* `name` - (Required) Name for the node.
* `output` - (Optional) Information about an output from the node. See [`definition.node.output` Block](#definitionnodeoutput-block) for details.
* `type` - (Required) Type of node. This value must match the name of the key you provide in `configuration`. Valid values: `Agent`, `Collector`, `Condition`, `InlineCode`, `Input`, `Iterator`, `KnowledgeBase`, `LambdaFunction`, `Lex`, `Output`, `Prompt`, `Retrieval`, `Storage`.

### `definition.node.configuration` Block

The `definition.node.configuration` configuration block supports the following arguments:

* `agent` - (Optional) Configurations for an agent node in your flow. Invokes an alias of an agent and returns the response. See [`definition.node.configuration.agent` Block](#definitionnodeconfigurationagent-block) for details.
* `collector` - (Optional) Configurations for a collector node in your flow. Collects an iteration of inputs and consolidates them into an array of outputs. This block has no arguments.
* `condition` - (Optional) Configurations for a Condition node in your flow. Sets conditions that lead to different branches of the flow. See [`definition.node.configuration.condition` Block](#definitionnodeconfigurationcondition-block) for details.
* `inline_code` - (Optional) Configurations for an inline code node in your flow. See [`definition.node.configuration.inline_code` Block](#definitionnodeconfigurationinline_code-block) for details.
* `input` - (Optional) Configurations for an input flow node in your flow. The node `inputs` can't be specified for this node. This block has no arguments.
* `iterator` - (Optional) Configurations for an iterator node in your flow. Takes an input that is an array and iteratively sends each item of the array as an output to the following node. The size of the array is also returned in the output. The output flow node at the end of the flow iteration returns a response for each member of the array. To return only one response, you can include a collector node downstream from the iterator node. This block has no arguments.
* `knowledge_base` - (Optional) Configurations for a knowledge base node in your flow. Queries a knowledge base and returns the retrieved results or generated response. See [`definition.node.configuration.knowledge_base` Block](#definitionnodeconfigurationknowledge_base-block) for details.
* `lambda_function` - (Optional) Configurations for a Lambda function node in your flow. Invokes a Lambda function. See [`definition.node.configuration.lambda_function` Block](#definitionnodeconfigurationlambda_function-block) for details.
* `lex` - (Optional) Configurations for a Lex node in your flow. Invokes an Amazon Lex bot to identify the intent of the input and return the intent as the output. See [`definition.node.configuration.lex` Block](#definitionnodeconfigurationlex-block) for details.
* `output` - (Optional) Configurations for an output flow node in your flow. The node `outputs` can't be specified for this node. This block has no arguments.
* `prompt` - (Optional) Configurations for a prompt node in your flow. Runs a prompt and generates the model response as the output. You can use a prompt from Prompt management or you can configure one in this node. See [`definition.node.configuration.prompt` Block](#definitionnodeconfigurationprompt-block) for details.
* `retrieval` - (Optional) Configurations for a Retrieval node in your flow. Retrieves data from an Amazon S3 location and returns it as the output. See [`definition.node.configuration.retrieval` Block](#definitionnodeconfigurationretrieval-block) for details.
* `storage` - (Optional) Configurations for a Storage node in your flow. Stores an input in an Amazon S3 location. See [`definition.node.configuration.storage` Block](#definitionnodeconfigurationstorage-block) for details.

### `definition.node.configuration.agent` Block

The `definition.node.configuration.agent` configuration block supports the following arguments:

* `agent_alias_arn` - (Required) ARN of the alias of the agent to invoke.

### `definition.node.configuration.condition` Block

The `definition.node.configuration.condition` configuration block supports the following arguments:

* `condition` - (Optional) List of conditions. See [`definition.node.configuration.condition.condition` Block](#definitionnodeconfigurationconditioncondition-block) for details.

### `definition.node.configuration.condition.condition` Block

The `definition.node.configuration.condition.condition` configuration block supports the following arguments:

* `expression` - (Optional) Expression that sets the condition. You must refer to at least one of the inputs in the condition. For more information, expand the Condition node section in [Node types in prompt flows](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-how-it-works.html#flows-nodes).
* `name` - (Required) Name for the condition that you can reference.

### `definition.node.configuration.inline_code` Block

The `definition.node.configuration.inline_code` configuration block supports the following arguments:

* `code` - (Required) Code that's executed in your inline code node.
* `language` - (Required) Programming language used by your inline code node.

### `definition.node.configuration.knowledge_base` Block

The `definition.node.configuration.knowledge_base` configuration block supports the following arguments:

* `guardrail_configuration` - (Optional) Configuration of a guardrail for knowledge base query and response generation. See [`definition.node.configuration.knowledge_base.guardrail_configuration` Block](#definitionnodeconfigurationknowledge_baseguardrail_configuration-block) for details.
* `inference_configuration` - (Optional) Configuration of model inference for knowledge base query and response generation. See [`definition.node.configuration.knowledge_base.inference_configuration` Block](#definitionnodeconfigurationknowledge_baseinference_configuration-block) for details.
* `knowledge_base_id` - (Required) Unique identifier of the knowledge base to query.
* `model_id` - (Required) Unique identifier of the model or inference profile to use to generate a response from the query results. Omit this field to return the retrieved results as an array.
* `number_of_results` - (Optional) Maximum number of results to retrieve from the knowledge base. Valid values are between 1 and 100.

### `definition.node.configuration.knowledge_base.guardrail_configuration` Block

The `definition.node.configuration.knowledge_base.guardrail_configuration` configuration block supports the following arguments:

* `guardrail_identifier` - (Required) Unique identifier of the guardrail.
* `guardrail_version` - (Required) Version of the guardrail.

### `definition.node.configuration.knowledge_base.inference_configuration` Block

The `definition.node.configuration.knowledge_base.inference_configuration` configuration block supports the following arguments:

* `text` - (Optional) Inference configurations for a text prompt. See [`definition.node.configuration.knowledge_base.inference_configuration.text` Block](#definitionnodeconfigurationknowledge_baseinference_configurationtext-block) for details.

### `definition.node.configuration.knowledge_base.inference_configuration.text` Block

The `definition.node.configuration.knowledge_base.inference_configuration.text` configuration block supports the following arguments:

* `max_tokens` - (Optional) Maximum number of tokens to return in the response.
* `stop_sequences` - (Optional) List of strings that define sequences after which the model stops generating.
* `temperature` - (Optional) Controls the randomness of the response. Choose a lower value for more predictable outputs and a higher value for more surprising outputs.
* `top_p` - (Optional) Percentage of most-likely candidates that the model considers for the next token.

### `definition.node.configuration.lambda_function` Block

The `definition.node.configuration.lambda_function` configuration block supports the following arguments:

* `lambda_arn` - (Required) ARN of the Lambda function to invoke.

### `definition.node.configuration.lex` Block

The `definition.node.configuration.lex` configuration block supports the following arguments:

* `bot_alias_arn` - (Required) ARN of the Amazon Lex bot alias to invoke.
* `locale_id` - (Required) Region to invoke the Amazon Lex bot in.

### `definition.node.configuration.prompt` Block

The `definition.node.configuration.prompt` configuration block supports the following arguments:

* `guardrail_configuration` - (Optional) Configuration of a guardrail for prompt generation. See [`definition.node.configuration.prompt.guardrail_configuration` Block](#definitionnodeconfigurationpromptguardrail_configuration-block) for details.
* `source_configuration` - (Required) Configuration of the prompt source, either inline or from Prompt management. See [`definition.node.configuration.prompt.source_configuration` Block](#definitionnodeconfigurationpromptsource_configuration-block) for details.

### `definition.node.configuration.prompt.guardrail_configuration` Block

The `definition.node.configuration.prompt.guardrail_configuration` configuration block supports the following arguments:

* `guardrail_identifier` - (Required) Unique identifier of the guardrail.
* `guardrail_version` - (Required) Version of the guardrail.

### `definition.node.configuration.prompt.source_configuration` Block

The `definition.node.configuration.prompt.source_configuration` configuration block supports the following arguments:

* `inline` - (Optional) Configurations for a prompt that is defined inline. See [`definition.node.configuration.prompt.source_configuration.inline` Block](#definitionnodeconfigurationpromptsource_configurationinline-block) for details.
* `resource` - (Optional) Configurations for a prompt from Prompt management. See [`definition.node.configuration.prompt.source_configuration.resource` Block](#definitionnodeconfigurationpromptsource_configurationresource-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline` Block

The `definition.node.configuration.prompt.source_configuration.inline` configuration block supports the following arguments:

* `additional_model_request_fields` - (Optional) Additional fields to be included in the model request for the Prompt node.
* `inference_configuration` - (Optional) Inference configurations for the prompt. See [`definition.node.configuration.prompt.source_configuration.inline.inference_configuration` Block](#definitionnodeconfigurationpromptsource_configurationinlineinference_configuration-block) for details.
* `model_id` - (Required) Unique identifier of the model or [inference profile](https://docs.aws.amazon.com/bedrock/latest/userguide/cross-region-inference.html) to run inference with.
* `template_configuration` - (Required) Prompt and variables in the prompt that can be replaced with values at runtime. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configuration-block) for details.
* `template_type` - (Required) Type of prompt template. Valid values: `TEXT`, `CHAT`.

### `definition.node.configuration.prompt.source_configuration.inline.inference_configuration` Block

The `definition.node.configuration.prompt.source_configuration.inline.inference_configuration` configuration block supports the following arguments:

* `text` - (Optional) Inference configurations for a text prompt. See [`definition.node.configuration.prompt.source_configuration.inline.inference_configuration.text` Block](#definitionnodeconfigurationpromptsource_configurationinlineinference_configurationtext-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline.inference_configuration.text` Block

The `definition.node.configuration.prompt.source_configuration.inline.inference_configuration.text` configuration block supports the following arguments:

* `max_tokens` - (Optional) Maximum number of tokens to return in the response.
* `stop_sequences` - (Optional) List of strings that define sequences after which the model stops generating.
* `temperature` - (Optional) Controls the randomness of the response. Choose a lower value for more predictable outputs and a higher value for more surprising outputs.
* `top_p` - (Optional) Percentage of most-likely candidates that the model considers for the next token.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration` configuration block supports the following arguments:

* `chat` - (Optional) Configurations to use the prompt in a conversational format. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchat-block) for details.
* `text` - (Optional) Configurations for the text in a message for a prompt. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.text` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationtext-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat` configuration block supports the following arguments:

* `input_variable` - (Optional) Variables in the prompt template. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.input_variable` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchatinput_variable-block) for details.
* `message` - (Optional) Messages in the chat for the prompt. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchatmessage-block) for details.
* `system` - (Optional) System prompts that provide context to the model or describe how it should behave. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.system` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchatsystem-block) for details.
* `tool_configuration` - (Optional) Configuration information for the tools that the model can use when generating a response. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configuration-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.input_variable` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.input_variable` configuration block supports the following arguments:

* `name` - (Required) Name of the variable.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message` configuration block supports the following arguments:

* `content` - (Required) Content for the message you pass to, or receive from, a model. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message.content` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchatmessagecontent-block) for details.
* `role` - (Required) Role that the message belongs to.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message.content` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message.content` configuration block supports the following arguments:

* `cache_point` - (Optional) Cache checkpoint within a message. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message.content.cache_point` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchatmessagecontentcache_point-block) for details.
* `text` - (Optional) Text in the message.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message.content.cache_point` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.message.content.cache_point` configuration block supports the following arguments:

* `type` - (Required) Type of the cache point block. Valid values: `default`.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.system` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.system` configuration block supports the following arguments:

* `cache_point` - (Optional) Cache checkpoint within a tool designation. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.system.cache_point` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchatsystemcache_point-block) for details.
* `text` - (Optional) Text in the system prompt.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.system.cache_point` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.system.cache_point` configuration block supports the following arguments:

* `type` - (Required) Type of the cache point block. Valid values: `default`.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration` configuration block supports the following arguments:

* `tool` - (Optional) Tools to pass to a model. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configurationtool-block) for details.
* `tool_choice` - (Optional) Which tools the model should request when invoked. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool_choice` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configurationtool_choice-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool_choice` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool_choice` configuration block supports the following arguments:

* `any` - (Optional) Tools, at least one of which must be requested by the model. No text is generated but the results of tool use are sent back to the model to help generate a response. This block has no arguments.
* `auto` - (Optional) Tools. The model automatically decides whether to call a tool or to generate text instead. This block has no arguments.
* `tool` - (Optional) Specific tool that the model must request. No text is generated but the results of tool use are sent back to the model to help generate a response. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool_choice.tool` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configurationtool_choicetool-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool_choice.tool` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool_choice.tool` configuration block supports the following arguments:

* `name` - (Required) Name of the tool.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool` configuration block supports the following arguments:

* `cache_point` - (Optional) Cache checkpoint within a tool designation. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.cache_point` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configurationtoolcache_point-block) for details.
* `tool_spec` - (Optional) Specification for the tool. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.tool_spec` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configurationtooltool_spec-block) for details.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.cache_point` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.cache_point` configuration block supports the following arguments:

* `type` - (Required) Type of the cache point block. Valid values: `default`.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.tool_spec` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.tool_spec` configuration block supports the following arguments:

* `description` - (Optional) Description of the tool.
* `input_schema` - (Optional) Input schema of the tool. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.tool_spec.input_schema` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationchattool_configurationtooltool_specinput_schema-block) for details.
* `name` - (Required) Name of the tool.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.tool_spec.input_schema` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.chat.tool_configuration.tool.tool_spec.input_schema` configuration block supports the following arguments:

* `json` - (Optional) JSON object defining the input schema for the tool.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.text` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.text` configuration block supports the following arguments:

* `cache_point` - (Optional) Cache checkpoint within a template configuration. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.text.cache_point` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationtextcache_point-block) for details.
* `input_variable` - (Optional) Variables in the prompt template. See [`definition.node.configuration.prompt.source_configuration.inline.template_configuration.text.input_variable` Block](#definitionnodeconfigurationpromptsource_configurationinlinetemplate_configurationtextinput_variable-block) for details.
* `text` - (Required) Message for the prompt.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.text.cache_point` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.text.cache_point` configuration block supports the following arguments:

* `type` - (Required) Type of the cache point block. Valid values: `default`.

### `definition.node.configuration.prompt.source_configuration.inline.template_configuration.text.input_variable` Block

The `definition.node.configuration.prompt.source_configuration.inline.template_configuration.text.input_variable` configuration block supports the following arguments:

* `name` - (Required) Name of the variable.

### `definition.node.configuration.prompt.source_configuration.resource` Block

The `definition.node.configuration.prompt.source_configuration.resource` configuration block supports the following arguments:

* `prompt_arn` - (Required) ARN of the prompt from Prompt management.

### `definition.node.configuration.retrieval` Block

The `definition.node.configuration.retrieval` configuration block supports the following arguments:

* `service_configuration` - (Required) Configurations for the service to use for retrieving data to return as the output from the node. See [`definition.node.configuration.retrieval.service_configuration` Block](#definitionnodeconfigurationretrievalservice_configuration-block) for details.

### `definition.node.configuration.retrieval.service_configuration` Block

The `definition.node.configuration.retrieval.service_configuration` configuration block supports the following arguments:

* `s3` - (Optional) Configurations for the Amazon S3 location from which to retrieve data to return as the output from the node. See [`definition.node.configuration.retrieval.service_configuration.s3` Block](#definitionnodeconfigurationretrievalservice_configurations3-block) for details.

### `definition.node.configuration.retrieval.service_configuration.s3` Block

The `definition.node.configuration.retrieval.service_configuration.s3` configuration block supports the following arguments:

* `bucket_name` - (Required) Name of the Amazon S3 bucket from which to retrieve data.

### `definition.node.configuration.storage` Block

The `definition.node.configuration.storage` configuration block supports the following arguments:

* `service_configuration` - (Required) Configurations for the service to use for storing the input into the node. See [`definition.node.configuration.storage.service_configuration` Block](#definitionnodeconfigurationstorageservice_configuration-block) for details.

### `definition.node.configuration.storage.service_configuration` Block

The `definition.node.configuration.storage.service_configuration` configuration block supports the following arguments:

* `s3` - (Optional) Configurations for the Amazon S3 location in which to store the input into the node. See [`definition.node.configuration.storage.service_configuration.s3` Block](#definitionnodeconfigurationstorageservice_configurations3-block) for details.

### `definition.node.configuration.storage.service_configuration.s3` Block

The `definition.node.configuration.storage.service_configuration.s3` configuration block supports the following arguments:

* `bucket_name` - (Required) Name of the Amazon S3 bucket in which to store the input into the node.

### `definition.node.configuration.input` Block

The `definition.node.configuration.input` configuration block has no arguments.

### `definition.node.configuration.output` Block

The `definition.node.configuration.output` configuration block has no arguments.

### `definition.node.input` Block

The `definition.node.input` configuration block supports the following arguments:

* `category` - (Optional) How input data flows between iterations in a DoWhile loop.
* `expression` - (Required) Expression that formats the input for the node. For an explanation of how to create expressions, see [Expressions in Prompt flows in Amazon Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-expressions.html).
* `name` - (Required) Name for the input that you can reference.
* `type` - (Required) Data type of the input. If the input doesn't match this type at runtime, a validation error is thrown.

### `definition.node.output` Block

The `definition.node.output` configuration block supports the following arguments:

* `name` - (Required) Name for the output that you can reference.
* `type` - (Required) Data type of the output. If the output doesn't match this type at runtime, a validation error is thrown.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the flow.
* `created_at` - Time at which the flow was created.
* `id` - Unique identifier of the flow.
* `status` - Status of the flow.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time at which the flow was last updated.
* `version` - Version of the flow.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Agents Flow using the `id`. For example:

```terraform
import {
  to = aws_bedrockagent_flow.example
  id = "ABCDEFGHIJ"
}
```

Using `terraform import`, import Bedrock Agents Flow using the `id`. For example:

```console
% terraform import aws_bedrockagent_flow.example ABCDEFGHIJ
```

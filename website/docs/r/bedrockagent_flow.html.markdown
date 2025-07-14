---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_flow"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Flow.
---

# Resource: aws_bedrockagent_flow

Terraform resource for managing an AWS Bedrock Agents Flow.

### Basic Usage

```terraform
resource "aws_bedrockagent_flow" "example" {
  name               = "example-flow"
  execution_role_arn = aws_iam_role.example.arn
}
```

## Example Usage

The default definition:

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

* `name` - (Required) A name for the flow.
* `execution_role_arn` - (Required) The Amazon Resource Name (ARN) of the service role with permissions to create and manage a flow. For more information, see [Create a service role for flows in Amazon Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-permissions.html) in the Amazon Bedrock User Guide.

The following arguments are optional:

* `description` - (Optional) A description for the flow.
* `customer_encryption_key_arn` - (Optional) The Amazon Resource Name (ARN) of the KMS key to encrypt the flow.
* `definition` - (Optional) A definition of the nodes and connections between nodes in the flow. See [Definition](#definition) for more information.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Definition

* `connection` - (Optional) A list of connection definitions in the flow. See [Connection](#connection) for more information.
* `node` - (Optional) A list of node definitions in the flow. See [Node](#node) for more information.

### Connection

* `name` - (Required) A name for the connection that you can reference.
* `source` - (Required) The node that the connection starts at.
* `target` - (Required) The node that the connection ends at.
* `type` - (Required) Whether the source node that the connection begins from is a condition node `Conditional` or not `Data`.
* `configuration` - (Required) Configuration of the connection. See [Connection Configuration](#connection-configuration) for more information.

### Connection Configuration

* `data` - (Optional) The configuration of a connection originating from a node that isn’t a Condition node. See [Data Connection Configuration](#data-connection-configuration) for more information.
* `conditional` - (Optional) The configuration of a connection originating from a Condition node. See [Conditional Connection Configuration](#conditional-connection-configuration) for more information.

#### Data Connection Configuration

* `source_output` - (Required) The name of the output in the source node that the connection begins from.
* `target_input` - (Required) The name of the input in the target node that the connection ends at.

#### Conditional Connection Configuration

* `condition` - (Required) The condition that triggers this connection. For more information about how to write conditions, see the Condition node type in the [Node types](https://docs.aws.amazon.com/bedrock/latest/userguide/node-types.html) topic in the Amazon Bedrock User Guide.

### Node

* `name` - (Required) A name for the node.
* `type` - (Required) The type of node. This value must match the name of the key that you provide in the configuration. Valid values: `Agent`, `Collector`, `Condition`, `Input`, `Iterator`, `KnowledgeBase`, `LambdaFunction`, `Lex`, `Output`, `Prompt`, `Retrieval`, `Storage`
* `configuration` - (Required) Contains configurations for the node. See [Node Configuration](#node-configuration) for more information.
* `input` - (Optional) A list of objects containing information about an input into the node. See [Node Input](#node-input) for more information.
* `output` - (Optional) A list of objects containing information about an output from the node. See [Node Output](#node-output) for more information.

### Node Input

* `name` - (Required) A name for the input that you can reference.
* `type` - (Required) The data type of the input. If the input doesn’t match this type at runtime, a validation error will be thrown.
* `expression` - (Required) An expression that formats the input for the node. For an explanation of how to create expressions, see [Expressions in Prompt flows in Amazon Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-expressions.html).
* `category` - (Optional) How input data flows between iterations in a DoWhile loop.

### Node Output

* `name` - (Required) A name for the output that you can reference.
* `type` - (Required) The data type of the output. If the output doesn’t match this type at runtime, a validation error will be thrown.

### Node Configuration

* `agent` - (Optional) Contains configurations for an agent node in your flow. Invokes an alias of an agent and returns the response. See [Agent Node Configuration](#agent-node-configuration) for more information.
* `collector` - (Optional) Contains configurations for a collector node in your flow. Collects an iteration of inputs and consolidates them into an array of outputs. This object has no fields.
* `condition` - (Optional) Contains configurations for a Condition node in your flow. Defines conditions that lead to different branches of the flow. See [Condition Node Configuration](#condition-node-configuration) for more information.
* `inline_code` - (Optional) Contains configurations for an inline code node in your flow. See [Inline Code Node Configuration](#inline-code-node-configuration) for more information.
* `input` - (Optional) Contains configurations for an input flow node in your flow. The node `inputs` can’t be specified for this node. This object has no fields.
* `iterator` - (Optional) Contains configurations for an iterator node in your flow. Takes an input that is an array and iteratively sends each item of the array as an output to the following node. The size of the array is also returned in the output. The output flow node at the end of the flow iteration will return a response for each member of the array. To return only one response, you can include a collector node downstream from the iterator node. This object has no fields.
* `knowledge_base` - (Optional) Contains configurations for a knowledge base node in your flow. Queries a knowledge base and returns the retrieved results or generated response. See [Knowledge Base Node Configuration](#knowledge-base-node-configuration) for more information.
* `lambda_function` - (Optional) Contains configurations for a Lambda function node in your flow. Invokes a Lambda function. See [Lambda Function Node Configuration](#lambda-function-node-configuration) for more information.
* `lex` - (Optional) Contains configurations for a Lex node in your flow. Invokes an Amazon Lex bot to identify the intent of the input and return the intent as the output. See [Lex Node Configuration](#lex-node-configuration) for more information.
* `output` - (Optional) Contains configurations for an output flow node in your flow. The node `outputs` can’t be specified for this node. This object has no fields.
* `prompt` - (Optional) Contains configurations for a prompt node in your flow. Runs a prompt and generates the model response as the output. You can use a prompt from Prompt management or you can configure one in this node. See [Prompt Node Configuration](#prompt-node-configuration) for more information.
* `retrieval` - (Optional) Contains configurations for a Retrieval node in your flow. Retrieves data from an Amazon S3 location and returns it as the output. See [Retrieval Node Configuration](#retrieval-node-configuration) for more information.
* `storage` - (Optional) Contains configurations for a Storage node in your flow. Stores an input in an Amazon S3 location. See [Storage Node Configuration](#storage-node-configuration) for more information.

### Agent Node Configuration

* `agent_alias_arn` - (Required) The Amazon Resource Name (ARN) of the alias of the agent to invoke.

### Condition Node Configuration

* `condition` - (Optional) A list of conditions. See [Condition Config](#condition-config) for more information.

#### Condition Config

* `name` - (Required) A name for the condition that you can reference.
* `expression` - (Optional) Defines the condition. You must refer to at least one of the inputs in the condition. For more information, expand the Condition node section in [Node types in prompt flows](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-how-it-works.html#flows-nodes).

### Inline Code Node Configuration

* `code` - (Required) The code that's executed in your inline code node.
* `language` - (Required) The programming language used by your inline code node.

### Knowledge Base Node Configuration

* `knowledge_base_id` - (Required) The unique identifier of the knowledge base to query.
* `model_id` - (Required) The unique identifier of the model or inference profile to use to generate a response from the query results. Omit this field if you want to return the retrieved results as an array.
* `guardrail_configuration` - (Required) Contains configurations for a guardrail to apply during query and response generation for the knowledge base in this configuration. See [Guardrail Configuration](#guardrail-configuration) for more information.

#### Guardrail Configuration

* `guardrail_identifier` - (Required) The unique identifier of the guardrail.
* `guardrail_version` - (Required) The version of the guardrail.

### Lambda Function Node Configuration

* `lambda_arn` - (Required) The Amazon Resource Name (ARN) of the Lambda function to invoke.

### Lex Node Configuration

* `bot_alias_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon Lex bot alias to invoke.
* `locale_id` - (Required) The Region to invoke the Amazon Lex bot in

### Prompt Node Configuration

* `resource` - (Optional) Contains configurations for a prompt from Prompt management. See [Prompt Resource Configuration](#prompt-resource-configuration) for more information.
* `inline` - (Optional) Contains configurations for a prompt that is defined inline. See [Prompt Inline Configuration](#prompt-inline-configuration) for more information.

#### Prompt Resource Configuration

* `prompt_arn` - (Required) The Amazon Resource Name (ARN) of the prompt from Prompt management.

#### Prompt Inline Configuration

* `additional_model_request_fields` - (Optional) Additional fields to be included in the model request for the Prompt node.
* `inference_configuration` - (Optional) Contains inference configurations for the prompt. See [Prompt Inference Configuration](#prompt-inference-configuration) for more information.
* `model_id` - (Required) The unique identifier of the model or [inference profile](https://docs.aws.amazon.com/bedrock/latest/userguide/cross-region-inference.html) to run inference with.
* `template_type` - (Required) The type of prompt template. Valid values: `TEXT`, `CHAT`.
* `template_configuration` - (Required) Contains a prompt and variables in the prompt that can be replaced with values at runtime. See [Prompt Template Configuration](#prompt-template-configuration) for more information.

#### Prompt Inference Configuration

* `text` - (Optional) Contains inference configurations for a text prompt. See [Text Inference Configuration](#text-inference-configuration) for more information.

#### Text Inference Configuration

* `max_tokens` - (Optional) Maximum number of tokens to return in the response.
* `stop_sequences` - (Optional) List of strings that define sequences after which the model will stop generating.
* `temperature` - (Optional) Controls the randomness of the response. Choose a lower value for more predictable outputs and a higher value for more surprising outputs.
* `top_p` - (Optional) Percentage of most-likely candidates that the model considers for the next token.

#### Prompt Template Configuration

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

### Retrieval Node Configuration

* `service_configuration` - (Required) Contains configurations for the service to use for retrieving data to return as the output from the node. See [Retrieval Service Configuration](#retrieval-service-configuration) for more information.

#### Retrieval Service Configuration

* `s3` - (Optional) Contains configurations for the Amazon S3 location from which to retrieve data to return as the output from the node. See [Retrieval S3 Service Configuration](#retrieval-s3-service-configuration) for more information.

#### Retrieval S3 Service Configuration

* `bucket_name` - (Required) The name of the Amazon S3 bucket from which to retrieve data.

### Storage Node Configuration

* `service_configuration` - (Required) Contains configurations for a Storage node in your flow. Stores an input in an Amazon S3 location. See [Storage Service Configuration](#storage-service-configuration) for more information.

#### Storage Service Configuration

* `s3` - (Optional) Contains configurations for the service to use for storing the input into the node. See [Storage S3 Service Configuration](#storage-s3-service-configuration) for more information.

#### Storage S3 Service Configuration

* `bucket_name` - (Required) The name of the Amazon S3 bucket in which to store the input into the node.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the flow.
* `id` - The unique identifier of the flow.
* `created_at` - The time at which the flow was created.
* `updated_at` - The time at which the flow was last updated.
* `version` - The version of the flow.
* `status` - The status of the flow.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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

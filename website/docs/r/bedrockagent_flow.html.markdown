---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_flow"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Flows.
---

# Resource: aws_bedrockagent_flow

Terraform resource for managing an AWS Bedrock Agents Flow.

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

### Basic Usage

```terraform
resource "aws_bedrockagent_flow" "example" {
  name               = "example-flow"
  execution_role_arn = aws_iam_role.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) A name for the flow.
* `execution_role_arn` - (Required) The Amazon Resource Name (ARN) of the service role with permissions to create and manage a flow. For more information, see [Create a service role for flows in Amazon Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-permissions.html) in the Amazon Bedrock User Guide.

The following arguments are optional:

* `description` - (Optional) A description for the flow.
* `customer_encryption_key_arn` - (Optional) The Amazon Resource Name (ARN) of the KMS key to encrypt the flow.
* `definition` - (Optional) A definition of the nodes and connections between nodes in the flow. See [Definition Config](#definition-config) for more information.
* `tags` (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Definition Config

* `connection` - (Optional) A list of connection definitions in the flow. See [Connection Config](#connection-config) for more information.
* `node` - (Optional) A list of node definitions in the flow. See [Node Config](#node-config) for more information.

### Connection Config

* `name` - (Required) A name for the connection that you can reference.
* `source` - (Required) The node that the connection starts at.
* `target` - (Required) The node that the connection ends at.
* `type` - (Required) Whether the source node that the connection begins from is a condition node `Conditional` or not `Data`.
* `configuration` - (Required) Configuration of the connection. See [Connection Configuration](#connection-configuration) for more information.

### Connection Configuration

* `data` - (Optional) The configuration of a connection originating from a node that isn’t a Condition node. See [Data Connection Configuration](#data-connection-configuration) for more information.
* `conditional` - (Optional) The configuration of a connection originating from a Condition node. See [Conditional Connection Configuration](#conditional-connection-configuration) for more information.

### Data Connection Configuration

* `source_output` - (Required) The name of the output in the source node that the connection begins from.
* `target_input` - (Required) The name of the input in the target node that the connection ends at.

### Conditional Connection Configuration

* `condition` - (Required) The condition that triggers this connection. For more information about how to write conditions, see the Condition node type in the [Node types](https://docs.aws.amazon.com/bedrock/latest/userguide/node-types.html) topic in the Amazon Bedrock User Guide.

### Node Config

* `name` - (Required) A name for the node.
* `type` - (Required) The type of node. This value must match the name of the key that you provide in the configuration. Valid values: `Agent`, `Collector`, `Condition`, `Input`, `Iterator`, `KnowledgeBase`, `LambdaFunction`, `Lex`, `Output`, `Prompt`, `Retrieval`, `Storage`
* `configuration` - (Required) Contains configurations for the node. See [Node Configuration](#node-configuration) for more information.
* `inputs` - (Optional) A list of objects containing information about an input into the node. See [Input Config](#input-config) for more information.
* `outputs` - (Optional) A list of containing information about an output from the node. See [Output Config](#output-config) for more information.

### Input Config

* `name` - (Required) A name for the input that you can reference.
* `type` - (Required) The data type of the input. If the input doesn’t match this type at runtime, a validation error will be thrown.
* `expression` - (Required) An expression that formats the input for the node. For an explanation of how to create expressions, see [Expressions in Prompt flows in Amazon Bedrock](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-expressions.html).

### Output Config
* `name` - (Required) A name for the output that you can reference.
* `type` - (Required) The data type of the output. If the output doesn’t match this type at runtime, a validation error will be thrown.

### Node Configuration

* `agent` - (Optional) Contains configurations for an agent node in your flow. Invokes an alias of an agent and returns the response. See [Agent Node Configuration](#agent-node-configuration) for more information.
* `collector` - (Optional) Contains configurations for a collector node in your flow. Collects an iteration of inputs and consolidates them into an array of outputs. This object has no fields.
* `condition` - (Optional) Contains configurations for a Condition node in your flow. Defines conditions that lead to different branches of the flow. See [Condition Node Configuration](#condition-node-configuration) for more information.
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

* `conditions` - (Optional) A list of conditions. See [Condition Config](#condition-config) for more information.

#### Condition Config

* `name` - (Required) A name for the condition that you can reference.
* `expression` - (Optional) Defines the condition. You must refer to at least one of the inputs in the condition. For more information, expand the Condition node section in [Node types in prompt flows](https://docs.aws.amazon.com/bedrock/latest/userguide/flows-how-it-works.html#flows-nodes).

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

TODO

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

* `arn` - ARN of the Flow.
* `id` - ID of the Flow.

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

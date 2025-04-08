---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_flow"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Flows.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
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

* `name` - (Required) Name of the Flow.
* `execution_role_arn` - (Required) ARN of the role.

The following arguments are optional:

* `description` - (Optional) Description of the Flow.
* `customer_encryption_key_arn` - (Optional) Customer encryption key ARN.
* `definition` - (Optional) Definition of the Flow. See [Definition Config](#definition-config) for more information.
* `tags` (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Definition Config

* `connection` - (Optional) List of connections. See [Connection Config](#connection-config) for more information
* `node` - (Optional) List of nodes. See [Node Config](#node-config) for more information

### Connection Config

* `name` - (Required) Name of the connection.
* `source` - (Required) Source of the connection.
* `target` - (Required) Target of the connection.
* `type` - (Required) Type of the connection, which is either `Data` or `Conditional`
* `configuration` - (Required) Configuration of the connection. See [Connection Configuration Config](#connection-configuration-config) for more information.

### Connection Configuration Config

* `data` - (Optional) Data connection configuration. See [Data Connection Configuration Config](#data-connection-configuration-config) for more information.
* `conditional` - (Optional) Conditional connection configuration. See [Conditional Connection Configuration Config](#conditional-connection-configuration-config) for more information.

### Data Connection Configuration Config

* `source_output` - (Required) The source output field name.
* `target_input` - (Required) The target input field name.

### Conditional Connection Configuration Config

* `condition` - (Required) The connection condition name.

### Node Config

TODO

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

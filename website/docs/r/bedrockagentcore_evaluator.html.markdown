---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_evaluator"
description: |-
  Manages an AWS Bedrock AgentCore Evaluator.
---

# Resource: aws_bedrockagentcore_evaluator

Manages an AWS Bedrock AgentCore Evaluator. An evaluator scores how an agent performs. You can configure it in one of two ways: an LLM-as-a-Judge evaluator that uses a model to score agent behavior against your instructions and a rating scale, or a code-based evaluator that runs a Lambda function you provide.

## Example Usage

### LLM-as-a-Judge with Numerical Rating Scale

```terraform
resource "aws_bedrockagentcore_evaluator" "example" {
  evaluator_name = "helpfulness_evaluator"
  description    = "Rates assistant helpfulness from 1 to 5"
  level          = "TRACE"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Given the {context} and the {assistant_turn}, compare against {expected_response} and rate from 1 to 5."

      rating_scale {
        numerical {
          definition = "Not helpful at all."
          value      = 1
          label      = "1"
        }
        numerical {
          definition = "Extremely helpful."
          value      = 5
          label      = "5"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "us.amazon.nova-2-lite-v1:0"

          inference_config {
            max_tokens  = 1024
            temperature = 0
            top_p       = 1
          }
        }
      }
    }
  }
}
```

### LLM-as-a-Judge with Categorical Rating Scale

```terraform
resource "aws_bedrockagentcore_evaluator" "example" {
  evaluator_name = "tone_evaluator"
  level          = "SESSION"

  evaluator_config {
    llm_as_a_judge {
      instructions = "Classify the tone of the {assistant_turn} given the {context}."

      rating_scale {
        categorical {
          definition = "Friendly, helpful tone."
          label      = "POSITIVE"
        }
        categorical {
          definition = "Neutral or terse tone."
          label      = "NEUTRAL"
        }
        categorical {
          definition = "Unhelpful or dismissive tone."
          label      = "NEGATIVE"
        }
      }

      model_config {
        bedrock_evaluator_model_config {
          model_id = "us.amazon.nova-2-lite-v1:0"
        }
      }
    }
  }
}
```

### Code-based Evaluator (Lambda)

```terraform
resource "aws_bedrockagentcore_evaluator" "example" {
  evaluator_name = "lambda_evaluator"
  level          = "TOOL_CALL"

  evaluator_config {
    code_based {
      lambda_config {
        lambda_arn                = aws_lambda_function.example.arn
        lambda_timeout_in_seconds = 60
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `evaluator_config` - (Required) Configuration that defines how the evaluator assesses agent performance. See [`evaluator_config`](#evaluator_config-block) below.
* `evaluator_name` - (Required, Forces new resource) Name of the evaluator. Must match the pattern `^[a-zA-Z][a-zA-Z0-9_]{0,47}$`.
* `level` - (Required) Evaluation level that determines the scope of evaluation. Valid values: `TOOL_CALL`, `TRACE`, `SESSION`.

The following arguments are optional:

* `description` - (Optional) Description of the evaluator. Length 1–200.
* `kms_key_arn` - (Optional) ARN of a customer-managed KMS key used to encrypt the evaluator's sensitive data. Only symmetric encryption keys are supported.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `evaluator_config` Block

Exactly one of `llm_as_a_judge` or `code_based` must be specified.

* `code_based` - (Optional) Configuration that runs a Lambda function you provide to score the agent. See [`code_based`](#code_based-block) below.
* `llm_as_a_judge` - (Optional) Configuration that uses a Bedrock model to score the agent. See [`llm_as_a_judge`](#llm_as_a_judge-block) below.

### `llm_as_a_judge` Block

* `instructions` - (Required) Instructions that tell the model how to score the agent.
* `model_config` - (Required) Which Bedrock model to use. See [`model_config`](#model_config-block) below.
* `rating_scale` - (Required) Scale used to score the agent. See [`rating_scale`](#rating_scale-block) below.

### `rating_scale` Block

Exactly one of `numerical` or `categorical` must be specified.

* `categorical` - (Optional) One or more categorical rating scale definitions. See [`categorical`](#categorical-block) below.
* `numerical` - (Optional) One or more numerical rating scale definitions. See [`numerical`](#numerical-block) below.

### `numerical` Block

* `definition` - (Required) Description that explains what this numerical rating represents.
* `label` - (Required) Label for this numerical rating option. Length 1–100.
* `value` - (Required) Numerical value for this rating option. Must be at least 0.

### `categorical` Block

* `definition` - (Required) Description that explains what this categorical rating represents.
* `label` - (Required) Label for this categorical rating option. Length 1–100.

### `model_config` Block

* `bedrock_evaluator_model_config` - (Required) Amazon Bedrock model configuration. See [`bedrock_evaluator_model_config`](#bedrock_evaluator_model_config-block) below.

### `bedrock_evaluator_model_config` Block

* `additional_model_request_fields` - (Optional) JSON-encoded model-specific request fields, for settings not covered by `inference_config`.
* `inference_config` - (Optional) Settings that control how the model generates its response. See [`inference_config`](#inference_config-block) below.
* `model_id` - (Required) Identifier of the Amazon Bedrock model to use for evaluation.

### `inference_config` Block

* `max_tokens` - (Optional) Maximum number of tokens to generate in the model response. Must be at least 1.
* `stop_sequences` - (Optional) List of sequences that cause the model to stop generating tokens.
* `temperature` - (Optional) Temperature value that controls randomness. Range 0–1.
* `top_p` - (Optional) Top-p sampling parameter. Range 0–1.

### `code_based` Block

* `lambda_config` - (Required) Lambda function configuration. See [`lambda_config`](#lambda_config-block) below.

### `lambda_config` Block

* `lambda_arn` - (Required) ARN of the Lambda function that runs the evaluation.
* `lambda_timeout_in_seconds` - (Optional) Time in seconds to wait for the Lambda function before timing out. Defaults to 60. Range 1–300.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Timestamp when the evaluator was created.
* `evaluator_arn` - ARN of the evaluator.
* `evaluator_id` - Unique identifier of the evaluator.
* `locked_for_modification` - Whether the evaluator is locked because it is in use by an active online evaluation.
* `status` - Current status of the evaluator.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute to import Bedrock AgentCore Evaluator. For example:

```terraform
import {
  to = aws_bedrockagentcore_evaluator.example
  identity = {
    evaluator_id = "helpfulness_evaluator-abc1234567"
  }
}
```

### Identity Schema

#### Required

* `evaluator_id` (String) Unique identifier of the evaluator.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Evaluator using the evaluator ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_evaluator.example
  id = "helpfulness_evaluator-abc1234567"
}
```

Using `terraform import`, import Bedrock AgentCore Evaluator using the evaluator ID. For example:

```console
% terraform import aws_bedrockagentcore_evaluator.example helpfulness_evaluator-abc1234567
```

---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_memory_strategy"
description: |-
  Manages an AWS Bedrock AgentCore Memory Strategy.
---

# Resource: aws_bedrockagentcore_memory_strategy

Manages an AWS Bedrock AgentCore Memory Strategy. Memory strategies define how the agent processes and organizes information within a memory, such as semantic understanding, summarization, or custom processing logic.

**Important Limitations:**

- Each memory can have a maximum of 6 strategies total
- Only one strategy of each built-in type (`SEMANTIC`, `SUMMARIZATION`, `USER_PREFERENCE`, `EPISODIC`) can exist per memory
- Multiple `CUSTOM` strategies are allowed (subject to the total limit of 6)

## Example Usage

### Semantic Strategy

```terraform
resource "aws_bedrockagentcore_memory_strategy" "semantic" {
  name        = "semantic-strategy"
  memory_id   = aws_bedrockagentcore_memory.example.id
  type        = "SEMANTIC"
  description = "Semantic understanding strategy"
  namespaces  = ["default"]
}
```

### Summarization Strategy

```terraform
resource "aws_bedrockagentcore_memory_strategy" "summary" {
  name        = "summary-strategy"
  memory_id   = aws_bedrockagentcore_memory.example.id
  type        = "SUMMARIZATION"
  description = "Text summarization strategy"
  namespaces  = ["{sessionId}"]
}
```

### User Preference Strategy

```terraform
resource "aws_bedrockagentcore_memory_strategy" "user_pref" {
  name        = "user-preference-strategy"
  memory_id   = aws_bedrockagentcore_memory.example.id
  type        = "USER_PREFERENCE"
  description = "User preference tracking strategy"
  namespaces  = ["preferences"]
}
```

### Episodic Strategy

```terraform
resource "aws_bedrockagentcore_memory_strategy" "episodic" {
  name        = "episodic-strategy"
  memory_id   = aws_bedrockagentcore_memory.example.id
  type        = "EPISODIC"
  description = "Episodic memory strategy"
  namespaces  = ["/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"]
}
```

### Custom Strategy with Semantic Override

```terraform
resource "aws_bedrockagentcore_memory_strategy" "custom_semantic" {
  name                      = "custom-semantic-strategy"
  memory_id                 = aws_bedrockagentcore_memory.example.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.example.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Custom semantic processing strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = "SEMANTIC_OVERRIDE"

    consolidation {
      append_to_prompt = "Focus on extracting key semantic relationships and concepts"
      model_id         = "anthropic.claude-3-sonnet-20240229-v1:0"
    }

    extraction {
      append_to_prompt = "Extract and categorize semantic information"
      model_id         = "anthropic.claude-3-haiku-20240307-v1:0"
    }
  }
}
```

### Custom Strategy with Summary Override

```terraform
resource "aws_bedrockagentcore_memory_strategy" "custom_summary" {
  name        = "custom-summary-strategy"
  memory_id   = aws_bedrockagentcore_memory.example.id
  type        = "CUSTOM"
  description = "Custom summarization strategy"
  namespaces  = ["summaries"]

  configuration {
    type = "SUMMARY_OVERRIDE"

    consolidation {
      append_to_prompt = "Create concise summaries while preserving key details"
      model_id         = "anthropic.claude-3-sonnet-20240229-v1:0"
    }
  }
}
```

### Custom Strategy with User Preference Override

```terraform
resource "aws_bedrockagentcore_memory_strategy" "custom_user_pref" {
  name        = "custom-user-preference-strategy"
  memory_id   = aws_bedrockagentcore_memory.example.id
  type        = "CUSTOM"
  description = "Custom user preference tracking strategy"
  namespaces  = ["user_prefs"]

  configuration {
    type = "USER_PREFERENCE_OVERRIDE"

    consolidation {
      append_to_prompt = "Consolidate user preferences and behavioral patterns"
      model_id         = "anthropic.claude-3-sonnet-20240229-v1:0"
    }

    extraction {
      append_to_prompt = "Extract user preferences and interaction patterns"
      model_id         = "anthropic.claude-3-haiku-20240307-v1:0"
    }
  }
}
```

### Custom Strategy with Episodic Override

```terraform
resource "aws_bedrockagentcore_memory_strategy" "custom_episodic" {
  name                      = "custom-episodic-strategy"
  memory_id                 = aws_bedrockagentcore_memory.example.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.example.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Custom episodic processing strategy"
  namespaces                = ["/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"]

  configuration {
    type = "EPISODIC_OVERRIDE"

    consolidation {
      append_to_prompt = "Consolidate episodic memories into coherent narratives"
      model_id         = "anthropic.claude-3-sonnet-20240229-v1:0"
    }

    extraction {
      append_to_prompt = "Extract key events and episodes from interactions"
      model_id         = "anthropic.claude-3-haiku-20240307-v1:0"
    }
  }
}
```

### Custom Strategy with Self-Managed Configuration

```terraform
resource "aws_bedrockagentcore_memory_strategy" "self_managed" {
  name                      = "self-managed-strategy"
  memory_id                 = aws_bedrockagentcore_memory.example.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.example.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Self-managed processing strategy"

  configuration {
    type = "SELF_MANAGED"

    self_managed {
      historical_context_window_size = 10

      trigger_conditions = {
        message_based_trigger = {
          message_count = 12
        }
      }

      invocation_configuration {
        topic_arn                    = aws_sns_topic.example.arn
        payload_delivery_bucket_name = aws_s3_bucket.example.bucket
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the memory strategy. Changing this forces a new resource, because the service API does not support renaming a strategy.
* `memory_id` - (Required) ID of the memory to associate with this strategy. Changing this forces a new resource.
* `type` - (Required) Type of memory strategy. Valid values: `SEMANTIC`, `SUMMARIZATION`, `USER_PREFERENCE`, `EPISODIC`, `CUSTOM`. Changing this forces a new resource. Note that only one strategy of each built-in type (`SEMANTIC`, `SUMMARIZATION`, `USER_PREFERENCE`, `EPISODIC`) can exist per memory.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the memory strategy. Once set, a description cannot be removed via update because the service API ignores a null description and retains the previously stored value.
* `configuration` - (Optional) Custom configuration block. Required when `type` is `CUSTOM`, must be omitted for other types. See [`configuration`](#configuration) below.
* `namespaces` - (Optional) Set of namespace identifiers where this strategy applies. Required for all strategies except `CUSTOM` strategies using `SELF_MANAGED` configuration, where the service rejects namespaces and this argument must be omitted.

### `configuration`

The `configuration` block supports the following:

* `type` - (Required) Type of custom override. Valid values: `SEMANTIC_OVERRIDE`, `SUMMARY_OVERRIDE`, `USER_PREFERENCE_OVERRIDE`, `EPISODIC_OVERRIDE`, `SELF_MANAGED`. Changing this forces a new resource.
* `consolidation` - (Optional) Consolidation configuration for processing and organizing memory content. See [`consolidation`](#consolidation) below. Once added, this block cannot be removed without recreating the resource. Cannot be used with `type` set to `SELF_MANAGED`.
* `extraction` - (Optional) Extraction configuration for identifying and extracting relevant information. See [`extraction`](#extraction) below. Cannot be used with `type` set to `SUMMARY_OVERRIDE` or `SELF_MANAGED`. Once added, this block cannot be removed without recreating the resource.
* `self_managed` - (Optional) Self-managed processing configuration. Required when `type` is `SELF_MANAGED` and only valid for that type. See [`self_managed`](#self_managed) below.

### `self_managed`

The `self_managed` block supports the following:

* `invocation_configuration` - (Required) Configuration used to invoke the self-managed memory processing pipeline. See [`invocation_configuration`](#invocation_configuration) below.
* `historical_context_window_size` - (Optional) Number of historical messages to include in processing context. Valid range: `0` to `50`. Defaults to `4`.
* `trigger_conditions` - (Optional) Conditions that trigger memory processing. See [`trigger_conditions`](#trigger_conditions) below. When omitted, the service supplies the documented defaults for all three trigger types.

### `invocation_configuration`

The `invocation_configuration` block supports the following:

* `topic_arn` - (Required) ARN of the SNS topic for job notifications.
* `payload_delivery_bucket_name` - (Required) S3 bucket name for event payload delivery.

### `trigger_conditions`

The `trigger_conditions` object supports any combination of the following:

* `message_based_trigger` - (Optional) Message-based condition. Its `message_count` argument accepts values from `1` to `50` and defaults to `6`.
* `token_based_trigger` - (Optional) Token-based condition. Its `token_count` argument accepts values from `100` to `500000` and defaults to `5000`.
* `time_based_trigger` - (Optional) Idle-time condition. Its `idle_session_timeout` argument accepts values from `10` to `3000` seconds and defaults to `20`.

The service returns all three normalized conditions even when only a subset is configured. Unconfigured conditions are retained in state as computed values.

### `consolidation`

The `consolidation` block supports the following:

* `append_to_prompt` - (Required) Additional text to append to the model prompt for consolidation processing.
* `model_id` - (Required) ID of the foundation model to use for consolidation processing.

### `extraction`

The `extraction` block supports the following:

* `append_to_prompt` - (Required) Additional text to append to the model prompt for extraction processing.
* `model_id` - (Required) ID of the foundation model to use for extraction processing.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `memory_strategy_id` - Unique identifier of the Memory Strategy. This corresponds to the service `strategyId` identifier (AWS API / CloudFormation terminology).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Memory Strategy using the `memory_id,strategy_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_memory_strategy.example
  id = "MEMORY1234567890,STRATEGY0987654321"
}
```

Using `terraform import`, import Bedrock AgentCore Memory Strategy using the `memory_id,strategy_id`. For example:

```console
% terraform import aws_bedrockagentcore_memory_strategy.example MEMORY1234567890,STRATEGY0987654321
```

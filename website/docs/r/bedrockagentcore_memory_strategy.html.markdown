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
- Only one strategy of each built-in type (`SEMANTIC`, `SUMMARIZATION`, `USER_PREFERENCE`) can exist per memory
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

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the memory strategy.
* `memory_id` - (Required) ID of the memory to associate with this strategy. Changing this forces a new resource.
* `type` - (Required) Type of memory strategy. Valid values: `SEMANTIC`, `SUMMARIZATION`, `USER_PREFERENCE`, `CUSTOM`. Changing this forces a new resource. Note that only one strategy of each built-in type (`SEMANTIC`, `SUMMARIZATION`, `USER_PREFERENCE`) can exist per memory.
* `namespaces` - (Required) Set of namespace identifiers where this strategy applies. Namespaces help organize and scope memory content.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the memory strategy.
* `configuration` - (Optional) Custom configuration block. Required when `type` is `CUSTOM`, must be omitted for other types. See [`configuration`](#configuration) below.

### `configuration`

The `configuration` block supports the following:

* `type` - (Required) Type of custom override. Valid values: `SEMANTIC_OVERRIDE`, `SUMMARY_OVERRIDE`, `USER_PREFERENCE_OVERRIDE`. Changing this forces a new resource.
* `consolidation` - (Optional) Consolidation configuration for processing and organizing memory content. See [`consolidation`](#consolidation) below. Once added, this block cannot be removed without recreating the resource.
* `extraction` - (Optional) Extraction configuration for identifying and extracting relevant information. See [`extraction`](#extraction) below. Cannot be used with `type` set to `SUMMARY_OVERRIDE`. Once added, this block cannot be removed without recreating the resource.

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

* `id` - Unique identifier of the Memory Strategy.

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

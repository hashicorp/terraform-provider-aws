---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_prepare"
description: |-
  Prepares an Amazon Bedrock Agent for use.
---

# Action: aws_bedrockagent_agent_prepare

~> **Note:** `aws_bedrockagent_agent_prepare` is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Prepares an Amazon Bedrock Agent for use. This action creates a DRAFT version of the agent that contains the latest changes.

For information about Amazon Bedrock Agents, see the [Amazon Bedrock User Guide](https://docs.aws.amazon.com/bedrock/latest/userguide/agents.html). For specific information about preparing agents, see the [PrepareAgent](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_agent_PrepareAgent.html) page in the Amazon Bedrock API Reference.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagent_agent" "example" {
  agent_name              = "example-agent"
  agent_resource_role_arn = aws_iam_role.agent.arn
  foundation_model        = "anthropic.claude-v2"
  instruction             = "You are a helpful assistant."
  prepare_agent           = false
}

action "aws_bedrockagent_agent_prepare" "example" {
  config {
    agent_id = aws_bedrockagent_agent.example.agent_id
  }
}

resource "terraform_data" "prepare_trigger" {
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_bedrockagent_agent_prepare.example]
    }
  }

  depends_on = [aws_bedrockagent_agent.example]
}
```

### Prepare After Action Group Changes

```terraform
resource "aws_bedrockagent_agent_action_group" "example" {
  agent_id          = aws_bedrockagent_agent.example.agent_id
  action_group_name = "example-action-group"
  prepare_agent     = false
  action_group_executor {
    lambda = aws_lambda_function.example.function_name
  }
}

action "aws_bedrockagent_agent_prepare" "after_changes" {
  config {
    agent_id = aws_bedrockagent_agent.example.agent_id
  }
}

resource "terraform_data" "prepare_after_changes" {
  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.aws_bedrockagent_agent_prepare.after_changes]
    }
  }

  depends_on = [aws_bedrockagent_agent_action_group.example]
}
```

## Argument Reference

This action supports the following arguments:

* `agent_id` - (Required) Unique identifier of the agent to prepare.
* `region` - (Optional) Region where this action should be [run](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

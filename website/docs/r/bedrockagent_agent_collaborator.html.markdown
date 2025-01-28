---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_collaborator"
description: |-
  Terraform resource for managing an AWS Bedrock Agents Agent Collaborator.
---
# Resource: aws_bedrockagent_agent_collaborator

Terraform resource for managing an AWS Bedrock Agents Agent Collaborator.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "example_agent_trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["bedrock.amazonaws.com"]
      type        = "Service"
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
    condition {
      test     = "ArnLike"
      values   = ["arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "example_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-3-5-sonnet-20241022-v2:0",
    ]
  }
  statement {
    actions = ["bedrock:GetAgentAlias", "bedrock:InvokeAgent"]
    resources = [
      "arn:${data.aws_partition.current_agent.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*",
      "arn:${data.aws_partition.current_agent.partition}:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent-alias/*"
    ]
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_agent_trust.json
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_"
}

resource "aws_iam_role_policy" "example" {
  policy = data.aws_iam_policy_document.example_agent_permissions.json
  role   = aws_iam_role.example.id
}

resource "aws_bedrockagent_agent" "example_collaborator" {
  agent_name                  = "my-agent-collaborator"
  agent_resource_role_arn     = aws_iam_role.example.arn
  idle_session_ttl_in_seconds = 500
  foundation_model            = "anthropic.claude-3-5-sonnet-20241022-v2:0"
  instruction                 = "do what the supervisor tells you to do"
}

resource "aws_bedrockagent_agent" "example_supervisor" {
  agent_name                  = "my-agent-supervisor"
  agent_resource_role_arn     = aws_iam_role.example.arn
  agent_collaboration         = "SUPERVISOR"
  idle_session_ttl_in_seconds = 500
  foundation_model            = "anthropic.claude-3-5-sonnet-20241022-v2:0"
  instruction                 = "tell the sub agent what to do"
  prepare_agent               = false
}

resource "aws_bedrockagent_agent_alias" "example" {
  agent_alias_name = "my-agent-alias"
  agent_id         = aws_bedrockagent_agent.example_collaborator.agent_id
  description      = "Test Alias"
}

resource "aws_bedrockagent_agent_collaborator" "example" {
  agent_id                   = aws_bedrockagent_agent.example_supervisor.agent_id
  collaboration_instruction  = "tell the other agent what to do"
  collaborator_name          = "my-collab-example"
  relay_conversation_history = "TO_COLLABORATOR"

  agent_descriptor {
    alias_arn = aws_bedrockagent_agent_alias.example.agent_alias_arn
  }
}
```

## Argument Reference

The following arguments are required:

* `agent_id` - (Required) ID if the agent to associate the collaborator.
* `collaboration_instruction` - (Required) Instruction to give the collaborator.
* `collbaorator_name` - (Required) Name of this collaborator.

The following arguments are optional:

* `prepare_agent` (Optional) Whether to prepare the agent after creation or modification. Defaults to `true`.
* `relay_conversation_history` - (Optional) Configure relaying the history to the collaborator.

### `agent_descriptor` Block

The `agent_descriptor` configuration block supports the following arguments:

* `alias_arn` - (Required) ARN of the Alias of an Agent to use as the collaborator.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `collaborator_id` - ID of the Agent Collaborator.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Agents Agent Collaborator using a comma-delimited string combining `agent_id`, `agent_version`, and `collaborator_id`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_collaborator.example
  id = "9LSJO0BFI8,DRAFT,AG3TN4RQIY"
}
```

Using `terraform import`, import Bedrock Agents Agent Collaborator using a comma-delimited string combining `agent_id`, `agent_version`, and `collaborator_id`. For example:

```console
% terraform import aws_bedrockagent_agent_collaborator.example 9LSJO0BFI8,DRAFT,AG3TN4RQIY
```

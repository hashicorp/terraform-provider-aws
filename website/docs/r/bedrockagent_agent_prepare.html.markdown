---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_prepare"
description: |-
  Terraform resource for managing an AWS Bedrock Agent Prepare action.
---

# Resource: aws_bedrockagent_agent_prepare

Terraform resource for managing an AWS Bedrock Agent Prepare action.

The `aws_bedrockagent_agent_prepare` resource creates a `DRAFT` version of an existing Bedrock agent that can be used for internal testing. This is an action resource that triggers the preparation of an agent rather than managing a persistent AWS resource.

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
      values   = ["arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "example_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/anthropic.claude-v2",
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

resource "aws_bedrockagent_agent" "example" {
  agent_name                  = "my-agent-name"
  agent_resource_role_arn     = aws_iam_role.example.arn
  idle_session_ttl_in_seconds = 500
  foundation_model            = "anthropic.claude-v2"
}


resource "aws_bedrockagent_agent_prepare" "example" {
  id = aws_bedrockagent_agent.example.id
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) The unique identifier of the agent to prepare. Must be a 10-character alphanumeric string. Changing this value will trigger a replacement of the resource.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `prepared_at` - The timestamp when the agent was last prepared, in RFC3339 format.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

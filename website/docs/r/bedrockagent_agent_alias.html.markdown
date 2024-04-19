---
subcategory: "Agents for Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_alias"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Agent Alias.
---
# Resource: aws_bedrockagent_agent_alias

Terraform resource for managing an AWS Agents for Amazon Bedrock Agent Alias.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.example_agent_trust.json
  name_prefix        = "AmazonBedrockExecutionRoleForAgents_"
}

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
      values   = ["arn:aws:bedrock:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:agent/*"]
      variable = "AWS:SourceArn"
    }
  }
}

data "aws_iam_policy_document" "example_agent_permissions" {
  statement {
    actions = ["bedrock:InvokeModel"]
    resources = [
      "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-v2",
    ]
  }
}

resource "aws_iam_role_policy" "example" {
  policy = data.aws_iam_policy_document.example_agent_permissions.json
  role   = aws_iam_role.example.id
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_bedrockagent_agent" "test" {
  agent_name              = "my-agent-name"
  agent_resource_role_arn = aws_iam_role.example.arn
  idle_ttl                = 500
  foundation_model        = "anthropic.claude-v2"
}
resource "aws_bedrockagent_agent_alias" "example" {
  agent_alias_name = "my-agent-alias"
  agent_id         = aws_bedrockagent_agent.test.agent_id
  description      = "Test ALias"
}
```

## Argument Reference

The following arguments are required:

* `agent_alias_name` - (Required) Name of the alias.
* `agent_id` - (Required) Identifier of the agent to create an alias for.
* `tags` - (Optional) Key-value tags for the place index. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The following arguments are optional:

* `description` - (Optional) Description of the alias of the agent.
* `routing_configuration` - (Optional) Routing configuration of the alias

### routing_configuration

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

The following arguments are required:

* `agent_version` - (Required) Version of the agent the alias routes to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_alias_arn` - ARN of the Agent Alias.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Alias using the `ABCDE12345,FGHIJ67890`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_alias.example
  id = "ABCDE12345,FGHIJ67890"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Alias using the `AGENT_ID,ALIAS_ID`. For example:

```console
% terraform import aws_bedrockagent_agent_alias.example AGENT_ID,ALIAS_ID
```

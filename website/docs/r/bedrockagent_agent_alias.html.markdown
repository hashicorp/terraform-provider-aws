---
subcategory: "Bedrock Agents"
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
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.name}::foundation-model/anthropic.claude-v2",
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
  agent_name              = "my-agent-name"
  agent_resource_role_arn = aws_iam_role.example.arn
  idle_ttl                = 500
  foundation_model        = "anthropic.claude-v2"
}

resource "aws_bedrockagent_agent_alias" "example" {
  agent_alias_name = "my-agent-alias"
  agent_id         = aws_bedrockagent_agent.example.agent_id
  description      = "Test Alias"
}
```

## Argument Reference

The following arguments are required:

* `agent_alias_name` - (Required) Name of the alias.
* `agent_id` - (Required, Forces new resource) Identifier of the agent to create an alias for.

The following arguments are optional:

* `description` - (Optional) Description of the alias.
* `routing_configuration` - (Optional) Details about the routing configuration of the alias. See [`routing_configuration` Block](#routing_configuration-block) for details.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `routing_configuration` Block

The `routing_configuration` configuration block supports the following arguments:

* `agent_version` - (Optional) Version of the agent with which the alias is associated.
* `provisioned_throughput` - (Optional) ARN of the Provisioned Throughput assigned to the agent alias.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_alias_arn` - ARN of the alias.
* `agent_alias_id` - Unique identifier of the alias.
* `id` - Alias ID and agent ID separated by `,`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Alias using the alias ID and the agent ID separated by `,`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_alias.example
  id = "66IVY0GUTF,GGRRAED6JP"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Alias using the alias ID and the agent ID separated by `,`. For example:

```console
% terraform import aws_bedrockagent_agent_alias.example 66IVY0GUTF,GGRRAED6JP
```

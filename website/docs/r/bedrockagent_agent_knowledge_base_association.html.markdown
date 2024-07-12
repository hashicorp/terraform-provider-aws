---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_agent_knowledge_base_association"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Agent Knowledge Base Association.
---
# Resource: aws_bedrockagent_agent_knowledge_base_association

Terraform resource for managing an AWS Agents for Amazon Bedrock Agent Knowledge Base Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagent_agent_knowledge_base_association" "example" {
  agent_id             = "GGRRAED6JP"
  description          = "Example Knowledge base"
  knowledge_base_id    = "EMDPPAYPZI"
  knowledge_base_state = "ENABLED"
}
```

## Argument Reference

The following arguments are required:

* `agent_id` - (Required, Forces new resource) Unique identifier of the agent with which you want to associate the knowledge base.
* `description` - (Required) Description of what the agent should use the knowledge base for.
* `knowledge_base_id` - (Required, Forces new resource) Unique identifier of the knowledge base to associate with the agent.
* `knowledge_base_state` - (Required) Whether to use the knowledge base when sending an [InvokeAgent](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_agent-runtime_InvokeAgent.html) request. Valid values: `ENABLED`, `DISABLED`.

The following arguments are optional:

* `agent_version` - (Optional, Forces new resource) Version of the agent with which you want to associate the knowledge base. Valid values: `DRAFT`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Agent ID, agent version, and knowledge base ID separated by `,`.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Knowledge Base Association using the agent ID, the agent version, and the knowledge base ID separated by `,`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_knowledge_base_association.example
  id = "GGRRAED6JP,DRAFT,EMDPPAYPZI"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Knowledge Base Association using the agent ID, the agent version, and the knowledge base ID separated by `,`. For example:

```console
% terraform import aws_bedrockagent_agent_knowledge_base_association.example GGRRAED6JP,DRAFT,EMDPPAYPZI
```

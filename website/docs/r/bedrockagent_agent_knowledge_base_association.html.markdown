---
subcategory: "Agents for Amazon Bedrock"
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
  agent_id             = "012EXAMPLE"
  description          = "Example Knowledge base"
  knowledge_base_id    = "345EXAMPLE"
  knowledge_base_state = "ENABLED"
}
```

## Argument Reference

The following arguments are required:

* `agent_id` - (Required) The ID of the agent to associate.
* `description` - (Required) Description of the association.
* `knowledge_base_id` - (Required) The ID of the Knowledge Base to associate.
* `knowledge_base_state` - (Required) State of the association ENABLED or DISABLED.

The following arguments are optional:

* `agent_version` - (Optional) Agent version to associate the Knowledge Base to, currently only DRAFT.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Knowledge Base Association using the `012AGENTID-DRAFT-345678KBID`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_knowledge_base_association.example
  id = "012AGENTID-DRAFT-345678KBID"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Knowledge Base Association using the `012AGENTID-DRAFT-345678KBID`. For example:

```console
% terraform import aws_bedrockagent_agent_knowledge_base_association.example 012AGENTID-DRAFT-345678KBID
```

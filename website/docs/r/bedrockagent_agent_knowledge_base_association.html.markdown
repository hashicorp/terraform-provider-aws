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
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `tags` - (Optional) A map of tags assigned to the WorkSpaces Connection Alias. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Agent Knowledge Base Association. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Agent Knowledge Base Association using the `example_id_arg`. For example:

```terraform
import {
  to = aws_bedrockagent_agent_knowledge_base_association.example
  id = "agent_knowledge_base_association-id-12345678"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Agent Knowledge Base Association using the `example_id_arg`. For example:

```console
% terraform import aws_bedrockagent_agent_knowledge_base_association.example agent_knowledge_base_association-id-12345678
```

---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_agent_runtime_endpoint"
description: |-
  Manages an AWS Bedrock AgentCore Agent Runtime Endpoint.
---

# Resource: aws_bedrockagentcore_agent_runtime_endpoint

Manages an AWS Bedrock AgentCore Agent Runtime Endpoint. Agent Runtime Endpoints provide a network-accessible interface for interacting with agent runtimes, enabling external systems to communicate with and invoke agent capabilities.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_agent_runtime_endpoint" "example" {
  name             = "example-endpoint"
  agent_runtime_id = aws_bedrockagentcore_agent_runtime.example.agent_runtime_id
  description      = "Endpoint for agent runtime communication"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the agent runtime endpoint.
* `agent_runtime_id` - (Required) ID of the agent runtime this endpoint belongs to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `agent_runtime_version` - (Optional) Version of the agent runtime to use for this endpoint.
* `description` - (Optional) Description of the agent runtime endpoint.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_runtime_endpoint_arn` - ARN of the Agent Runtime Endpoint.
* `agent_runtime_arn` - ARN of the associated Agent Runtime.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Agent Runtime Endpoint using the `agent_runtime_id` and `name` separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_agent_runtime_endpoint.example
  id = "AGENTRUNTIME1234567890,example-endpoint"
}
```

Using `terraform import`, import Bedrock AgentCore Agent Runtime Endpoint using the `agent_runtime_id` and `name` separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_agent_runtime_endpoint.example AGENTRUNTIME1234567890,example-endpoint
```

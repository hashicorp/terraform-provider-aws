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
resource "aws_bedrockagentcore_agent_runtime" "example" {
  name        = "example-agent-runtime"
  description = "Example agent runtime"
}

resource "aws_bedrockagentcore_agent_runtime_endpoint" "example" {
  name                  = "example-endpoint"
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.example.id
  agent_runtime_version = "1.0.0"
  description           = "Endpoint for agent runtime communication"
}
```

### Agent Runtime Endpoint with Custom Configuration

```terraform
resource "aws_bedrockagentcore_agent_runtime" "example" {
  name        = "production-agent-runtime"
  description = "Production agent runtime"
}

resource "aws_bedrockagentcore_agent_runtime_endpoint" "example" {
  name                  = "production-endpoint"
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.example.id
  agent_runtime_version = aws_bedrockagentcore_agent_runtime.example.version
  description           = "Production endpoint with specific version"
  client_token          = "unique-client-token-123"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the agent runtime endpoint.
* `agent_runtime_id` - (Required) ID of the agent runtime this endpoint belongs to.
* `agent_runtime_version` - (Required) Version of the agent runtime to use for this endpoint.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the agent runtime endpoint.
* `client_token` - (Optional) Unique identifier for request idempotency. If not provided, one will be generated automatically.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Agent Runtime Endpoint.
* `agent_runtime_arn` - ARN of the associated Agent Runtime.
* `id` - Unique identifier of the Agent Runtime Endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Agent Runtime Endpoint using the `agent_runtime_id` and `endpoint_name` separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_agent_runtime_endpoint.example
  id = "AGENTRUNTIME1234567890,example-endpoint"
}
```

Using `terraform import`, import Bedrock AgentCore Agent Runtime Endpoint using the `agent_runtime_id` and `endpoint_name` separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_agent_runtime_endpoint.example AGENTRUNTIME1234567890,example-endpoint
```

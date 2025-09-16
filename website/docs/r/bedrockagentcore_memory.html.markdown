---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_memory"
description: |-
  Manages an AWS Bedrock AgentCore Memory.
---
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->

# Resource: aws_bedrockagentcore_memory

Manages an AWS Bedrock AgentCore Memory. Memory provides persistent storage for AI agent interactions, allowing agents to retain context across conversations and sessions.

## Example Usage

### Basic Usage

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "bedrock-agentcore-memory-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonBedrockAgentCoreMemoryBedrockModelInferenceExecutionRolePolicy"
}

resource "aws_bedrockagentcore_memory" "example" {
  name                  = "example-memory"
  event_expiry_duration = 30
}
```

### Memory with Custom Encryption and Role

```terraform
resource "aws_kms_key" "example" {
  description = "KMS key for Bedrock AgentCore Memory"
}

resource "aws_bedrockagentcore_memory" "example" {
  name                      = "example-memory"
  description               = "Memory for customer service agent"
  event_expiry_duration     = 60
  encryption_key_arn        = aws_kms_key.example.arn
  memory_execution_role_arn = aws_iam_role.example.arn
  client_token              = "unique-client-token"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the memory.
* `event_expiry_duration` - (Required) Number of minutes after which memory events expire. Must be a positive integer.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the memory.
* `encryption_key_arn` - (Optional) ARN of the KMS key used to encrypt the memory. If not provided, AWS managed encryption is used.
* `memory_execution_role_arn` - (Optional) ARN of the IAM role that the memory service assumes to perform operations. Required when using custom memory strategies with model processing.
* `client_token` - (Optional) Unique identifier for request idempotency. If not provided, one will be generated automatically.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Memory.
* `id` - Unique identifier of the Memory.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Memory using the memory ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_memory.example
  id = "MEMORY1234567890"
}
```

Using `terraform import`, import Bedrock AgentCore Memory using the memory ID. For example:

```console
% terraform import aws_bedrockagentcore_memory.example MEMORY1234567890
```

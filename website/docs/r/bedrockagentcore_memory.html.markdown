---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_memory"
description: |-
  Manages an AWS Bedrock AgentCore Memory.
---

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
  name                  = "example_memory"
  event_expiry_duration = 30
}
```

### Memory with Custom Encryption and Role

```terraform
resource "aws_kms_key" "example" {
  description = "KMS key for Bedrock AgentCore Memory"
}

resource "aws_bedrockagentcore_memory" "example" {
  name                      = "example_memory"
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
* `event_expiry_duration` - (Required) Number of days after which memory events expire. Must be a positive integer in the range of 7 to 365.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the memory.
* `encryption_key_arn` - (Optional) ARN of the KMS key used to encrypt the memory. If not provided, AWS managed encryption is used.
* `indexed_key` - (Optional) Metadata keys to index for filtering. Up to 10 entries. Changing this forces a new resource to be created. See [`indexed_key`](#indexed_key) below.
* `memory_execution_role_arn` - (Optional) ARN of the IAM role that the memory service assumes to perform operations. Required when using custom memory strategies with model processing.
* `stream_delivery_resources` - (Optional) Configuration for streaming memory record data to external resources. See [`stream_delivery_resources`](#stream_delivery_resources) below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### indexed_key

* `key` - (Required) Metadata key name to index.
* `type` - (Required) Data type of the indexed key. Valid values are `STRING`, `STRINGLIST`, and `NUMBER`.

### stream_delivery_resources

* `resource` - (Required) List of stream delivery resource configurations. See [`resource`](#resource) below.

### resource

* `kinesis` - (Optional) Kinesis Data Stream configuration. See [`kinesis`](#kinesis) below.

### kinesis

* `data_stream_arn` - (Required) ARN of the Kinesis Data Stream.
* `content_configuration` - (Required) Content configurations for stream delivery. See [`content_configuration`](#content_configuration) below.

### content_configuration

* `type` - (Required) Type of content to stream. Valid value is `MEMORY_RECORDS`.
* `level` - (Optional) Level of detail for streamed content. Valid values are `METADATA_ONLY` and `FULL_CONTENT`. Defaults to `METADATA_ONLY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Memory.
* `id` - Unique identifier of the Memory.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
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

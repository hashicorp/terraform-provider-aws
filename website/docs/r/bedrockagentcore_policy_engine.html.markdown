---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_policy_engine"
description: |-
  Manages an AWS Bedrock AgentCore Policy Engine.
---

# Resource: aws_bedrockagentcore_policy_engine

Manages an AWS Bedrock AgentCore Policy Engine. A Policy Engine provides authorization capabilities for AI agents, enabling fine-grained access control over the actions and resources an agent runtime is permitted to use.

-> **Note:** Once `description` is set, it cannot be removed without replacing the resource. The AWS API does not support clearing description on an existing policy engine.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_policy_engine" "example" {
  name = "example_policy_engine"
}
```

### With Description

```terraform
resource "aws_bedrockagentcore_policy_engine" "example" {
  name        = "example_policy_engine"
  description = "Policy engine for customer service agent"
}
```

### With Custom Encryption

```terraform
resource "aws_kms_key" "example" {
  description             = "KMS key for Bedrock AgentCore Policy Engine"
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_policy_engine" "example" {
  name               = "example_policy_engine"
  description        = "Policy engine for customer service agent"
  encryption_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the policy engine. Must start with a letter and contain only letters, numbers, and underscores. Maximum length of 48 characters.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the policy engine. Must be between 1 and 4096 characters. Once set, cannot be removed without replacing the resource.
* `encryption_key_arn` - (Optional, Forces new resource) ARN of the KMS key used to encrypt the policy engine. If not provided, AWS managed encryption is used.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier of the Policy Engine.
* `policy_engine_arn` - ARN of the Policy Engine.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Policy Engine using the policy engine ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_policy_engine.example
  id = "policy-engine-id-12345678"
}
```

Using `terraform import`, import Bedrock AgentCore Policy Engine using the policy engine ID. For example:

```console
% terraform import aws_bedrockagentcore_policy_engine.example policy-engine-id-12345678
```

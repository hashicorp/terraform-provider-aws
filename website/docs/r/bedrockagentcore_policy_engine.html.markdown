---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_policy_engine"
description: |-
  Manages an AWS Bedrock AgentCore Policy Engine.
---

# Resource: aws_bedrockagentcore_policy_engine

Manages an AWS Bedrock AgentCore Policy Engine. A Policy Engine controls what actions and resources an agent runtime can use.

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

* `description` - (Optional) Description of the policy engine.
* `encryption_key_arn` - (Optional, Forces new resource) ARN of the KMS key used to encrypt the policy engine. If not set, AWS uses an AWS managed key.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy_engine_arn` - ARN of the Policy Engine.
* `policy_engine_id` - Unique identifier of the Policy Engine.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_policy_engine.example
  identity = {
    policy_engine_id = "policy-engine-id-12345678"
  }
}

resource "aws_bedrockagentcore_policy_engine" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `policy_engine_id` (String) Policy engine ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Bedrock AgentCore Policy Engine by policy engine ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_policy_engine.example
  id = "policy-engine-id-12345678"
}
```

Using `terraform import`, import a Bedrock AgentCore Policy Engine by policy engine ID. For example:

```console
% terraform import aws_bedrockagentcore_policy_engine.example policy-engine-id-12345678
```

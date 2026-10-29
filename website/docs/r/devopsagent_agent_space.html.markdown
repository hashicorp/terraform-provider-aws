---
subcategory: "DevOps Agent"
layout: "aws"
page_title: "AWS: aws_devopsagent_agent_space"
description: |-
  Manages an AWS DevOps Agent Space.
---

# Resource: aws_devopsagent_agent_space

Manages an AWS DevOps Agent Space.

## Example Usage

### Basic Usage

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name = "my-agent-space"
}
```

### With Description and Tags

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name        = "my-agent-space"
  description = "An example agent space"

  tags = {
    Environment = "production"
  }
}
```

### With KMS Encryption

The KMS key must allow the calling identity and the DevOps Agent service principal to use it. Configure the key policy as described in the [AWS DevOps Agent encryption documentation](https://docs.aws.amazon.com/devopsagent/latest/userguide/aws-devops-agent-security-encryption-at-rest-for-devops-agent.html#step-2-set-the-key-policy) before using this example.

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name        = "my-agent-space"
  kms_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Agent Space.

The following arguments are optional:

* `description` - (Optional) Description of the Agent Space.
* `kms_key_arn` - (Optional) ARN of the AWS KMS key used to encrypt resources. Forces new resource.
* `locale` - (Optional) Locale for the Agent Space, which determines the language used in agent responses (e.g., `en`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_space_id` - Unique identifier of the Agent Space.
* `arn` - ARN of the Agent Space.
* `created_at` - Timestamp when the Agent Space was created.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Timestamp when the Agent Space was last updated.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_devopsagent_agent_space.example
  identity = {
    agent_space_id = "08975e27-f587-49ca-b7b8-2e1019783803"
  }
}

resource "aws_devopsagent_agent_space" "example" {
  name = "my-agent-space"
}
```

### Identity Schema

#### Required

* `agent_space_id` (String) Unique identifier of the Agent Space.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Agent Space using the `agent_space_id`. For example:

```terraform
import {
  to = aws_devopsagent_agent_space.example
  id = "08975e27-f587-49ca-b7b8-2e1019783803"
}
```

Using `terraform import`, import DevOps Agent Space using the `agent_space_id`. For example:

```console
% terraform import aws_devopsagent_agent_space.example 08975e27-f587-49ca-b7b8-2e1019783803
```

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

```terraform
resource "aws_devopsagent_agent_space" "example" {
  name        = "my-agent-space"
  kms_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the Agent Space.

The following arguments are optional:

* `description` - (Optional) The description of the Agent Space.
* `kms_key_arn` - (Optional, Forces new resource) The ARN of the AWS KMS key used to encrypt resources.
* `locale` - (Optional) The locale for the Agent Space, which determines the language used in agent responses (e.g., `en`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `agent_space_id` - The unique identifier of the Agent Space.
* `arn` - The Amazon Resource Name (ARN) of the Agent Space.
* `created_at` - The timestamp when the Agent Space was created.
* `id` - The unique identifier of the Agent Space.
* `updated_at` - The timestamp when the Agent Space was last updated.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DevOps Agent Space using the `agent_space_id`. For example:

```terraform
import {
  to = aws_devopsagent_agent_space.example
  id = "space-12345678"
}
```

Using `terraform import`, import DevOps Agent Space using the `agent_space_id`. For example:

```console
% terraform import aws_devopsagent_agent_space.example space-12345678
```

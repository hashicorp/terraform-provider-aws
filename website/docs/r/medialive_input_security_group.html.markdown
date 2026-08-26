---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_input_security_group"
description: |-
  Manages an AWS MediaLive Input Security Group.
---

# Resource: aws_medialive_input_security_group

Manages an AWS MediaLive Input Security Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_medialive_input_security_group" "example" {
  whitelist_rules {
    cidr = "10.0.0.8/32"
  }

  tags = {
    ENVIRONMENT = "prod"
  }
}
```

## Argument Reference

The following arguments are required:

* `whitelist_rules` - (Required) Whitelist rules. See [Whitelist Rules](#whitelist-rules) for more details.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Whitelist Rules

* `cidr` (Required) - The IPv4 CIDR that's whitelisted.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - InputSecurityGroup Id.
* `arn` - ARN of the InputSecurityGroup.
* `inputs` - The list of inputs currently using this InputSecurityGroup.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_medialive_input_security_group.example
  identity = {
    id = "123456"
  }
}

resource "aws_medialive_input_security_group" "example" {
  # Configuration omitted for brevity
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the Input Security Group.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MediaLive Input Security Group using the `id`. For example:

```terraform
import {
  to = aws_medialive_input_security_group.example
  id = "123456"
}
```

Using `terraform import`, import MediaLive Input Security Group using the `id`. For example:

```console
% terraform import aws_medialive_input_security_group.example 123456
```

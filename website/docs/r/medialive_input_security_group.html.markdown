---
subcategory: "Elemental MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_input_security_group"
description: |-
  Terraform resource for managing an AWS MediaLive InputSecurityGroup.
---

# Resource: aws_medialive_input_security_group

Terraform resource for managing an AWS MediaLive InputSecurityGroup.

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

* `tags` - (Optional) A map of tags to assign to the InputSecurityGroup. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Whitelist Rules

* `cidr` (Required) - The IPv4 CIDR that's whitelisted.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - InputSecurityGroup Id.
* `arn` - ARN of the InputSecurityGroup.
* `inputs` - The list of inputs currently using this InputSecurityGroup.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

MediaLive InputSecurityGroup can be imported using the `id`, e.g.,

```
$ terraform import aws_medialive_input_security_group.example 123456
```

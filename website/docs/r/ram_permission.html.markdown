---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_permission"
description: |-
  Manages an AWS RAM (Resource Access Manager) Permission.
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

# Resource: aws_ram_permission

Manages an AWS RAM (Resource Access Manager) Permission.

## Example Usage

### Basic Usage

```terraform
resource "aws_ram_permission" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Permission.
* `example_attribute` - Brief description of the attribute.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RAM (Resource Access Manager) Permission using the `example_id_arg`. For example:

```terraform
import {
  to = aws_ram_permission.example
  id = "permission-id-12345678"
}
```

Using `terraform import`, import RAM (Resource Access Manager) Permission using the `example_id_arg`. For example:

```console
% terraform import aws_ram_permission.example permission-id-12345678
```

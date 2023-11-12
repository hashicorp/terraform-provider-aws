---
subcategory: "Connect Cases"
layout: "aws"
page_title: "AWS: aws_connectcases_field":
description: |-
  Terraform resource for managing an Amazon Connect Cases Field.
---

# Resource: aws_connectcases_field

Terraform resource for managing an Amazon Connect Cases Field.
See the [Create Field](https://docs.aws.amazon.com/cases/latest/APIReference/API_CreateField.html) for more information.

## Example Usage

```terraform
resource "aws_connectcases_domain" "example" {
  name = "example"
}

resource "aws_connectcases_field" "example" {
  name        = "example"
  description = "example description of field"
  domain_id   = aws_connectcases_domain.example.domain_id
  type        = "Text"
}
```

## Argument Reference

The following arguments are required:

* `name` - The name of the field.
* `domain_id` - The unique identifier of the Cases domain.
* `type` - Defines the data type, some system constraints, and default display of the field. Valid Values: `Text` | `Number` | `Boolean` | `DateTime` | `SingleSelect` | `Url`.

The following arguments are optional:

* `description` - The description of the field.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `field_arn` - The Amazon Resource Name (ARN) of the field.
* `field_id` - The unique identifier of a field.
* `namespace` - The namespace of a field.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Cases Field using the resource `id`. For example:

```terraform
import {
  to = aws_connectcases_field.example
  id = "a78c466c-36c6-4222-9edb-60866392ed84"
}
```

Using `terraform import`, import Amazon Connect Cases Field using the resource `id`. For example:

```console
% terraform import aws_connectcases_field.example a78c466c-36c6-4222-9edb-60866392ed84
```

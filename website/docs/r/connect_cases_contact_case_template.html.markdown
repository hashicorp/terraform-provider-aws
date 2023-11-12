---
subcategory: "Connect Cases"
layout: "aws"
page_title: "AWS: aws_connectcases_template":
description: |-
  Terraform resource for managing an Amazon Connect Cases Template.
---

# Resource: aws_connectcases_template

Terraform resource for managing an Amazon Connect Cases Template.
See the [Create Template](https://docs.aws.amazon.com/cases/latest/APIReference/API_CreateTemplate.html) for more information.

## Example Usage

```terraform
resource "aws_connectcases_domain" "example" {
  name = "example"
}

resource "aws_connectcases_field" "example" {
  name        = "example-top-panel"
  description = "example description of field"
  domain_id   = aws_connectcases_domain.example.domain_id
  type        = "Text"
}

resource "aws_connectcases_field" "example_2" {
  name        = "example-more-info"
  description = "example description of field"
  domain_id   = aws_connectcases_domain.example.domain_id
  type        = "Text"
}

resource "aws_connectcases_template" "test" {
  name        = "example"
  description = "example description of template"
  domain_id   = aws_connectcases_domain.example.domain_id
  status      = "Inactive"

  layout_configuration {
    default_layout = aws_connectcases_layout.example.id
  }

  required_fields {
    field_id = aws_connectcases_field.example.field_id
  }
}

resource "aws_connectcases_layout" "example" {
  name      = "example"
  domain_id = aws_connectcases_domain.example.domain_id

  content {
    more_info {
      sections {
        name = "more_info_example"
        field_group {
          fields {
            id = aws_connectcases_field.example.field_id
          }
        }
      }
    }
    top_panel {
      sections {
        name = "top_panel_example"
        field_group {
          fields {
            id = aws_connectcases_field.example_2.field_id
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - The name for the template. It must be unique per domain.
* `domain_id` - The unique identifier of the Cases domain.

The following arguments are optional:

* `description` - The description of the field.
* `layout_configuration` - A block that specifies the configuration of layouts associated to the template. [Documented below](#layout_configuration).
* `required_fields` - A list of fields that must contain a value for a case to be successfully created with this template. [Documented below](#required_fields).
* `status` - The status of the template. Defaults to `Inactive`

### `layout_configuration`

The `layout_configuration` configuration block supports the following attributes:

* `default_layout` - Unique identifier of a layout.

### `required_fields`

The `required_fields` configuration block supports the following attributes:

* `field_id` - Unique identifier of a field.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `template_arn` - The Amazon Resource Name (ARN) of the newly created template.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Cases Field using the resource `id`. For example:

```terraform
import {
  to = aws_connectcases_template.example
  id = "d9621d53-719b-423e-a35c-7832f8fe37c7"
}
```

Using `terraform import`, import Amazon Connect Cases Template using the resource `id`. For example:

```console
% terraform import aws_connectcases_template.example d9621d53-719b-423e-a35c-7832f8fe37c7
```

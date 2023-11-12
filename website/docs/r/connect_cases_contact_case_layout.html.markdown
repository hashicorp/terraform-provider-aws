---
subcategory: "Connect Cases"
layout: "aws"
page_title: "AWS: aws_connectcases_layout":
description: |-
  Terraform resource for managing an Amazon Connect Cases Layout.
---

# Resource: aws_connectcases_layout

Terraform resource for managing an Amazon Connect Cases Layout.
See the [Create Layout](https://docs.aws.amazon.com/cases/latest/APIReference/API_CreateLayout.html) for more information.

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

* `name` - The name of the layout. It must be unique for the Cases domain.
* `content` - A block that specifies information about which fields will be present in the layout, and information about the order of the fields. [Documented below](#content).

### `content`

The `content` configuration block supports the following attributes:

* `more_info` - A block of sections in a tab of the page layout. [Documented below](#sections).
* `top_panel` - A block of sections in a panel of the page layout. [Documented below](#sections).

### `sections`

The `sections` configuration block supports the following attributes:

* `field_group` - A block that consists of a group of fields and associated properties. [Documented below](#field_group).

### `field_group`

The `field_group` configuration block supports the following attributes:

* `fields ` - A block that represents an ordered list containing field related information. [Documented below](#fields).
* `name ` - Name of the field group.

### `fields`

The `fields` configuration block supports the following attributes:

* `id ` - Unique identifier of a field.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `layout_arn` - The Amazon Resource Name (ARN) of the newly created layout.
* `layout_id` - The unique identifier of the layout.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Connect Cases Field using the resource `id`. For example:

```terraform
import {
  to = aws_connectcases_layout.example
  id = "d9621d53-719b-423e-a35c-7832f8fe37c7"
}
```

Using `terraform import`, import Amazon Connect Cases Layout using the resource `id`. For example:

```console
% terraform import aws_connectcases_layout.example d9621d53-719b-423e-a35c-7832f8fe37c7
```

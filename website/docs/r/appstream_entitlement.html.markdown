---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_entitlement"
description: |-
  Provides an AppStream entitlement
---

# Resource: aws_appstream_entitlement

Provides an AppStream entitlement. Entitlements control access to specific applications within an AppStream stack based on user attributes.

## Example Usage

```terraform
resource "aws_appstream_stack" "example" {
  name = "example-stack"
}

resource "aws_appstream_entitlement" "example" {
  name           = "example-entitlement"
  stack_name     = aws_appstream_stack.example.name
  description    = "Engineering team entitlement"
  app_visibility = "ASSOCIATED"

  attributes {
    name  = "roles"
    value = "engineering"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for the entitlement.
* `stack_name` - (Required) Name of the stack with which the entitlement is associated.
* `app_visibility` - (Required) Specifies whether all or only selected apps are entitled. Valid values are `ALL` or `ASSOCIATED`.
* `attributes` - (Required) Set of attributes associated with the entitlement. See [`attributes`](#attributes) below.

The following arguments are optional:

* `description` - (Optional) Description of the entitlement.

### `attributes`

* `name` - (Required) Name of the attribute.
* `value` - (Required) Value of the attribute.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_time` - Date and time, in UTC and extended RFC 3339 format, when the entitlement was created.
* `last_modified_time` - Date and time, in UTC and extended RFC 3339 format, when the entitlement was last modified.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appstream_entitlement` using the `stack_name` and `name` separated by a slash (`/`). For example:

```terraform
import {
  to = aws_appstream_entitlement.example
  id = "stack-name/entitlement-name"
}
```

Using `terraform import`, import `aws_appstream_entitlement` using the `stack_name` and `name` separated by a slash (`/`). For example:

```console
% terraform import aws_appstream_entitlement.example stack-name/entitlement-name
```

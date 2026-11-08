---
subcategory: "AppStream 2.0"
layout: "aws"
page_title: "AWS: aws_appstream_application_entitlement_association"
description: |-
  Manages an AppStream Application Entitlement Association
---

# Resource: aws_appstream_application_entitlement_association

Manages an AppStream Application Entitlement Association. This resource associates an application with an entitlement for an AppStream stack.

## Example Usage

```terraform
resource "aws_appstream_stack" "example" {
  name = "example-stack"
}

resource "aws_appstream_entitlement" "example" {
  name           = "example-entitlement"
  stack_name     = aws_appstream_stack.example.name
  app_visibility = "ASSOCIATED"

  attributes {
    name  = "department"
    value = "engineering"
  }
}

resource "aws_appstream_application" "example" {
  name         = "example-application"
  display_name = "Example Application"
  # ... other application configuration
}

resource "aws_appstream_application_entitlement_association" "example" {
  stack_name             = aws_appstream_stack.example.name
  entitlement_name       = aws_appstream_entitlement.example.name
  application_identifier = aws_appstream_application.example.id
}
```

## Argument Reference

The following arguments are required:

* `stack_name` - (Required) Name of the stack.
* `entitlement_name` - (Required) Name of the entitlement.
* `application_identifier` - (Required) Identifier of the application.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique ID of the application entitlement association, consisting of the `stack_name`, `entitlement_name`, and `application_identifier` separated by slashes (`/`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appstream_application_entitlement_association` using the `stack_name`, `entitlement_name`, and `application_identifier` separated by slashes (`/`). For example:

```terraform
import {
  to = aws_appstream_application_entitlement_association.example
  id = "stack-name/entitlement-name/application-identifier"
}
```

Using `terraform import`, import `aws_appstream_application_entitlement_association` using the `stack_name`, `entitlement_name`, and `application_identifier` separated by slashes (`/`). For example:

```console
% terraform import aws_appstream_application_entitlement_association.example stack-name/entitlement-name/application-identifier
```

---
subcategory: "CodeCatalyst"
layout: "aws"
page_title: "AWS: aws_codecatalyst_project"
description: |-
  Terraform resource for managing an AWS CodeCatalyst Project.
---

# Resource: aws_codecatalyst_project

Terraform resource for managing an AWS CodeCatalyst Project.

## Example Usage

### Basic Usage

```terraform
resource "aws_codecatalyst_project" "test" {
  space_name   = "myproject"
  display_name = "MyProject"
  description  = "My CodeCatalyst Project created using Terraform"
}
```

## Argument Reference

The following arguments are required:

* `space_name` - (Required) The name of the space.
* `display_name` - (Required) The friendly name of the project that will be displayed to users.

The following arguments are optional:

* `description` - (Optional) The description of the project. This description will be displayed to all users of the project. We recommend providing a brief description of the project and its intended purpose.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the project in the space.
* `name` - The name of the project in the space.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeCatalyst Project using the `id`. For example:

```terraform
import {
  to = aws_codecatalyst_project.example
  id = "project-id-12345678"
}
```

Using `terraform import`, import CodeCatalyst Project using the `id`. For example:

```console
% terraform import aws_codecatalyst_project.example project-id-12345678
```

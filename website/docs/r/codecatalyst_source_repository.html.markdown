---
subcategory: "CodeCatalyst"
layout: "aws"
page_title: "AWS: aws_codecatalyst_source_repository"
description: |-
  Terraform resource for managing an AWS CodeCatalyst Source Repository.
---

# Resource: aws_codecatalyst_source_repository

Terraform resource for managing an AWS CodeCatalyst Source Repository.

## Example Usage

### Basic Usage

```terraform
resource "aws_codecatalyst_source_repository" "example" {
  name         = "example-repo"
  project_name = "example-project"
  space_name   = "example-space"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the source repository. For more information about name requirements, see [Quotas for source repositories](https://docs.aws.amazon.com/codecatalyst/latest/userguide/source-quotas.html).
* `space_name` - (Required) The name of the CodeCatalyst space.
* `project_name` - (Required) The name of the project in the CodeCatalyst space.

The following arguments are optional:

* `description` - (Optional) The description of the project. This description will be displayed to all users of the project. We recommend providing a brief description of the project and its intended purpose.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeCatalyst Source Repository using the `example_id_arg`. For example:

```terraform
import {
  to = aws_codecatalyst_source_repository.example
  id = "source_repository-id-12345678"
}
```

Using `terraform import`, import CodeCatalyst Source Repository using the `example_id_arg`. For example:

```console
% terraform import aws_codecatalyst_source_repository.example source_repository-id-12345678
```

---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_environment"
description: |-
  Terraform resource for managing an AWS DataZone Environment.
---

# Resource: aws_datazone_environment

Terraform resource for managing an AWS DataZone Environment.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_environment" "example" {
  name                 = "example"
  account_identifier   = data.aws_caller_identity.test.account_id
  account_region       = data.aws_region.test.name
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  user_parameters {
    name  = "consumerGlueDbName"
    value = "consumer"
  }

  user_parameters {
    name  = "producerGlueDbName"
    value = "producer"
  }

  user_parameters {
    name  = "workgroupName"
    value = "workgroup"
  }
}
```

## Argument Reference

The following arguments are required:

* `account_identifier` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `account_identifier` - (Optional) The ID of the Amazon Web Services account where the environment exists
* `account_region` - (Optional) The Amazon Web Services region where the environment exists.
* `blueprint_identifier` - (Optional) The blueprint with which the environment is created.
* `descrioption` - (Optional) The description of the environment.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - The time the environment was created.
* `created_by` - The user who created the environment.
* `arn` - ARN of the Environment. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Environment using the `domain_identifier`,`id`. For example:

```terraform
import {
  to = aws_datazone_environment.example
  id = "environment-id-12345678"
}
```

Using `terraform import`, import DataZone Environment using the `example_id_arg`. For example:

```console
% terraform import aws_datazone_environment.example environment-id-12345678
```

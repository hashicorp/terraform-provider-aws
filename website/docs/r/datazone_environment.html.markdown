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

* `domain_identifier` - (Required) The ID of the domain where the environment exists.
* `name` - (Required) The name of the environment.
* `profile_identifier` - (Required) The ID of the profile with which the environment is created.
* `project_identifier` - (Required) The ID of the project where the environment exists.

The following arguments are optional:

* `account_identifier` - (Optional) The ID of the Amazon Web Services account where the environment exists
* `account_region` - (Optional) The Amazon Web Services region where the environment exists.
* `blueprint_identifier` - (Optional) The blueprint with which the environment is created.
* `description` - (Optional) The description of the environment.
* `glossary_terms` - (Optional) The business glossary terms that can be used in this environment.
* `user_parameters` - (Optional) The user parameters that are used in the environment. See [User Parameters](#user-parameters) for more information.

### User Parameters

* `name` - (Required) The name of an environment profile parameter.
* `value` - (Required) The value of an environment profile parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - The time the environment was created.
* `created_by` - The user who created the environment.
* `id` - The ID of the environment.
* `last_deployment` - The details of the last deployment of the environment.
* `provider_environment` - The provider of the environment.
* `provisioned_resource` - The provisioned resources of this environment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Environment using the `domain_identifier,id`. For example:

```terraform
import {
  to = aws_datazone_environment.example
  id = "dzd_d2i7tzk3tnjjf4,5vpywijpwryec0"
}
```

Using `terraform import`, import DataZone Environment using the `domain_idntifier,id`. For example:

```console
% terraform import aws_datazone_environment.example dzd_d2i7tzk3tnjjf4,5vpywijpwryec0
```

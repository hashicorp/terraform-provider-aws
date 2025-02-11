---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_environment_profile"
description: |-
  Terraform resource for managing an AWS DataZone Environment Profile.
---

# Resource: aws_datazone_environment_profile

Terraform resource for managing an AWS DataZone Environment Profile.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "domain_execution_role" {
  name = "example-name"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = "example-name"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

resource "aws_datazone_domain" "test" {
  name                  = "example-name"
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

resource "aws_security_group" "test" {
  name = "example-name"
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = "example-name"
  description         = "desc"
  skip_deletion_check = true
}

data "aws_caller_identity" "test" {}
data "aws_region" "test" {}

data "aws_datazone_environment_blueprint" "test" {
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}

resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  provisioning_role_arn    = aws_iam_role.domain_execution_role.arn
  enabled_regions          = [data.aws_region.test.name]
}

resource "aws_datazone_environment_profile" "test" {
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.name
  description                      = "description"
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  name                             = "example-name"
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
  user_parameters {
    name  = "consumerGlueDbName"
    value = "value"
  }
}
```

## Argument Reference

The following arguments are required:

* `aws_account_id` - (Required) -  Id of the AWS account being used.
* `aws_account_region` - (Required) -  Desired region for environment profile.
* `domain_identifier` - (Required) -  Domain Identifier for environment profile.
* `name` - (Required) -  Name of the environment profile.
* `environment_blueprint_identifier` - (Required) -  ID of the blueprint which the environment will be created with.
* `project_identifier` - (Required) -  Project identifier for environment profile.

The following arguments are optional:

* `description` - (Optional) Description of environment profile.
* `user_parameters` - (Optional) -  Array of user parameters of the environment profile with the following attributes:
    * `name` - (Required) -  Name of the environment profile parameter.
    * `value` - (Required) -  Value of the environment profile parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Creation time of environment profile.
* `created_by` - Creator of environment profile.
* `id` - ID of environment profile.
* `updated_at` - Time of last update to environment profile.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Environment Profile using a comma-delimited string combining `id` and `domain_identifier`. For example:

```terraform
import {
  to = aws_datazone_environment_profile.example
  id = "environment_profile-id-12345678,domain-id-12345678"
}
```

Using `terraform import`, import DataZone Environment Profile using a comma-delimited string combining `id` and `domain_identifier`. For example:

```console
% terraform import aws_datazone_environment_profile.example environment_profile-id-12345678,domain-id-12345678
```

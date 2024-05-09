---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_environment_blueprint_configuration"
description: |-
  Terraform resource for managing an AWS DataZone Environment Blueprint Configuration.
---

# Resource: aws_datazone_environment_blueprint_configuration

Terraform resource for managing an AWS DataZone Environment Blueprint Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_domain" "example" {
  name                  = "example_domain"
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

data "aws_datazone_environment_blueprint" "default_data_lake" {
  domain_id = aws_datazone_domain.example.id
  name      = "DefaultDataLake"
  managed   = true
}

resource "aws_datazone_environment_blueprint_configuration" "example" {
  domain_id                = aws_datazone_domain.example.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.default_data_lake.id
  enabled_regions          = ["us-east-1"]

  regional_parameters = {
    us-east-1 = {
      S3Location = "s3://my-amazon-datazone-bucket"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_id` - (Required) ID of the Domain.
* `environment_blueprint_id` - (Required) ID of the Environment Blueprint
* `enabled_regions` (Required) - Regions in which the blueprint is enabled

The following arguments are optional:

* `manage_access_role_arn` - (Optional) ARN of the manage access role with which this blueprint is created.
* `provisioning_role_arn` - (Optional) ARN of the provisioning role with which this blueprint is created.
* `regional_parameters` - (Optional) Parameters for each region in which the blueprint is enabled

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Environment Blueprint Configuration using the `domain_id` and `environment_blueprint_id`, separated by a `/`. For example:

```terraform
import {
  to = aws_datazone_environment_blueprint_configuration.example
  id = "domain-id-12345/environment-blueprint-id-54321"
}
```

Using `terraform import`, import DataZone Environment Blueprint Configuration using the `domain_id` and `environment_blueprint_id`, separated by a `/`. For example:

```console
% terraform import aws_datazone_environment_blueprint_configuration.example domain-id-12345/environment-blueprint-id-54321
```

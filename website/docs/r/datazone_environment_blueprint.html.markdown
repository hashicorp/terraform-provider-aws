---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_environment_blueprint"
description: |-
  Terraform resource for managing an AWS DataZone Environment Blueprint.
---

# Resource: aws_datazone_environment_blueprint

Terraform resource for managing an AWS DataZone Environment Blueprint.

Use this resource to register a custom blueprint in an Amazon DataZone (SageMaker Unified Studio) domain. A custom blueprint of category `TOOLING` lets you define the foundational tooling resources that are provisioned for projects, backed by a CloudFormation template that you own.

~> **Note:** After a blueprint is created it must be enabled with [`aws_datazone_environment_blueprint_configuration`](datazone_environment_blueprint_configuration.html) before it can be used by a project profile.

## Example Usage

### Basic Usage

```terraform
resource "aws_datazone_environment_blueprint" "example" {
  domain_identifier  = aws_datazone_domain.example.id
  name               = "CustomTooling"
  description        = "Org-standard Tooling foundation"
  blueprint_category = "TOOLING"

  provisioning_properties {
    cloud_formation {
      template_url = "https://${aws_s3_bucket.tooling.bucket_regional_domain_name}/${aws_s3_object.tooling.key}"
    }
  }
}
```

### With User Parameters

```terraform
resource "aws_datazone_environment_blueprint" "example" {
  domain_identifier  = aws_datazone_domain.example.id
  name               = "CustomTooling"
  blueprint_category = "TOOLING"

  provisioning_properties {
    cloud_formation {
      template_url = "https://${aws_s3_bucket.tooling.bucket_regional_domain_name}/${aws_s3_object.tooling.key}"
    }
  }

  user_parameters {
    key_name            = "costCenter"
    field_type          = "String"
    default_value       = "engineering"
    is_editable         = true
    is_optional         = true
    is_update_supported = true
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) Identifier of the domain in which the blueprint is created. Changing this forces a new resource to be created.
* `name` - (Required) Name of the blueprint. Changing this forces a new resource to be created.
* `provisioning_properties` - (Required) Provisioning properties of the blueprint. See [`provisioning_properties`](#provisioning_properties) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `blueprint_category` - (Optional) Category of the blueprint. The only valid value is `TOOLING`.
* `description` - (Optional) Description of the blueprint.
* `user_parameters` - (Optional) User parameters of the blueprint. See [`user_parameters`](#user_parameters) below.

### provisioning_properties

* `cloud_formation` - (Required) CloudFormation provisioning properties. See [`cloud_formation`](#cloud_formation) below.

### cloud_formation

* `template_url` - (Required) URL of the CloudFormation template used to provision the blueprint. The template must be stored in an S3 bucket in the same account as the domain, and the domain's execution role must be able to read it (for example via the `SageMakerStudioDomainExecutionRolePolicy` managed policy).

### user_parameters

* `key_name` - (Required) Key name of the parameter.
* `field_type` - (Required) Field type of the parameter.
* `default_value` - (Optional) Default value of the parameter.
* `description` - (Optional) Description of the parameter.
* `is_editable` - (Optional) Whether the parameter is editable.
* `is_optional` - (Optional) Whether the parameter is optional.
* `is_update_supported` - (Optional) Whether the parameter value can be updated after creation.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the blueprint.
* `created_at` - Timestamp of when the blueprint was created.
* `glossary_terms` - Glossary terms attached to the blueprint.
* `provider_name` - Provider of the blueprint.
* `updated_at` - Timestamp of when the blueprint was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `2m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_datazone_environment_blueprint.example
  identity = {
    domain_identifier = "dzd_54nakfrb46xfaa"
    id                = "4qd7g9diohnoob"
  }
}
```

### Identity Schema

#### Required

* `domain_identifier` - (String) Identifier of the domain in which the blueprint is created.
* `id` - (String) ID of the blueprint.

#### Optional

* `account_id` - (String) AWS Account where this resource is managed.
* `region` - (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Environment Blueprint using the `domain_identifier` and `id` separated by a `,`. For example:

```terraform
import {
  to = aws_datazone_environment_blueprint.example
  id = "dzd_54nakfrb46xfaa,4qd7g9diohnoob"
}
```

Using `terraform import`, import DataZone Environment Blueprint using the `domain_identifier` and `id` separated by a `,`. For example:

```console
% terraform import aws_datazone_environment_blueprint.example dzd_54nakfrb46xfaa,4qd7g9diohnoob
```

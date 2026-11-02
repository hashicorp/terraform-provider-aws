---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_migration_project"
description: |-
  Manages an AWS DMS (Database Migration) Migration Project.
---

# Resource: aws_dms_migration_project

Manages an AWS DMS (Database Migration) Migration Project. A migration project groups the instance profile and data providers used together in a DMS Schema Conversion or homogeneous data migration.

Create an [instance profile](/docs/providers/aws/r/dms_instance_profile.html) and [data providers](/docs/providers/aws/r/dms_data_provider.html) before creating a migration project.

## Example Usage

### Basic Usage

```terraform
resource "aws_dms_migration_project" "example" {
  instance_profile_arn = aws_dms_instance_profile.example.arn

  source_data_provider_descriptor {
    data_provider_arn = aws_dms_data_provider.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn = aws_dms_data_provider.target.arn
  }
}
```

### With Secrets Manager Credentials

```terraform
resource "aws_dms_migration_project" "example" {
  name                 = "example"
  description          = "Example migration project"
  instance_profile_arn = aws_dms_instance_profile.example.arn

  source_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.source.arn
    secrets_manager_access_role_arn = aws_iam_role.example.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.source.arn
  }

  target_data_provider_descriptor {
    data_provider_arn               = aws_dms_data_provider.target.arn
    secrets_manager_access_role_arn = aws_iam_role.example.arn
    secrets_manager_secret_id       = aws_secretsmanager_secret.target.arn
  }

  schema_conversion_application_attributes {
    s3_bucket_path     = "s3://example-bucket"
    s3_bucket_role_arn = aws_iam_role.example.arn
  }

  tags = {
    Environment = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_profile_arn` - (Required) ARN of the instance profile associated with the migration project.
* `source_data_provider_descriptor` - (Required) Information about the source data provider. See [`source_data_provider_descriptor` Block](#source_data_provider_descriptor-block) below.
* `target_data_provider_descriptor` - (Required) Information about the target data provider. See [`target_data_provider_descriptor` Block](#target_data_provider_descriptor-block) below.

The following arguments are optional:

* `description` - (Optional) User-friendly description of the migration project.
* `name` - (Optional) User-friendly name for the migration project.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `schema_conversion_application_attributes` - (Optional) Schema conversion application attributes, including the S3 bucket path and S3 role ARN. See [`schema_conversion_application_attributes` Block](#schema_conversion_application_attributes-block) below.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `transformation_rules` - (Optional) JSON string that specifies the transformation rules for the migration project. Homogeneous data migrations do not support transformation rules.

### `source_data_provider_descriptor` Block

The following arguments are required:

* `data_provider_arn` - (Required) ARN of the data provider.

The following arguments are optional:

* `secrets_manager_access_role_arn` - (Optional) ARN of the IAM role used to access AWS Secrets Manager.
* `secrets_manager_secret_id` - (Optional) Identifier of the Secrets Manager secret used to store access credentials for the data provider.

### `target_data_provider_descriptor` Block

The following arguments are required:

* `data_provider_arn` - (Required) ARN of the data provider.

The following arguments are optional:

* `secrets_manager_access_role_arn` - (Optional) ARN of the IAM role used to access AWS Secrets Manager.
* `secrets_manager_secret_id` - (Optional) Identifier of the Secrets Manager secret used to store access credentials for the data provider.

### `schema_conversion_application_attributes` Block

The following arguments are optional:

* `s3_bucket_path` - (Optional) S3 bucket path that the application uses for exporting assessment reports.
* `s3_bucket_role_arn` - (Optional) ARN of the IAM role the application uses to access its S3 bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the migration project.
* `creation_time` - Time the migration project was created, in RFC3339 format.
* `instance_profile_name` - Name of the associated instance profile.
* `source_data_provider_descriptor.data_provider_name` - Name of the source data provider.
* `target_data_provider_descriptor.data_provider_name` - Name of the target data provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dms_migration_project.example
  identity = {
    arn = "arn:aws:dms:us-east-1:123456789012:migration-project:EXAMPLEABCDEFGHIJKLMNOPQRS"
  }
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the migration project.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a DMS migration project using its full ARN. For example:

```terraform
import {
  to = aws_dms_migration_project.example
  id = "arn:aws:dms:us-east-1:123456789012:migration-project:EXAMPLEABCDEFGHIJKLMNOPQRS"
}
```

Using `terraform import`, import a DMS migration project using its full ARN. For example:

```console
% terraform import aws_dms_migration_project.example arn:aws:dms:us-east-1:123456789012:migration-project:EXAMPLEABCDEFGHIJKLMNOPQRS
```
